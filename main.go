package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"github.com/HotCodeGroup/warscript-utils/balancer"
	"github.com/HotCodeGroup/warscript-utils/logging"
	"github.com/HotCodeGroup/warscript-utils/middlewares"
	"github.com/HotCodeGroup/warscript-utils/models"
	"github.com/HotCodeGroup/warscript-utils/postgresql"
	"github.com/HotCodeGroup/warscript-utils/redis"

	"google.golang.org/grpc"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	consulapi "github.com/hashicorp/consul/api"
	vaultapi "github.com/hashicorp/vault/api"
)

var logger *logrus.Logger

//nolint: gocyclo
func main() {
	var err error
	logger, err = logging.NewLogger(os.Stdout, os.Getenv("LOGENTRIESRUS_TOKEN"))
	if err != nil {
		log.Printf("can not create logger: %s", err)
		return
	}

	consulConfig := consulapi.DefaultConfig()
	consulConfig.Address = os.Getenv("CONSUL_ADDR")
	consul, err := consulapi.NewClient(consulConfig)
	if err != nil {
		logger.Errorf("can not connect consul service: %s", err)
		return
	}

	httpPort, grpcPort, err := balancer.GetPorts("warscript-users/bounds", "warscript-users", consul)
	if err != nil {
		logger.Errorf("can not find empry port: %s", err)
		return
	}

	vaultConfig := vaultapi.DefaultConfig()
	vaultConfig.Address = os.Getenv("VAULT_ADDR")
	vault, err := vaultapi.NewClient(vaultConfig)
	if err != nil {
		logger.Errorf("can not connect vault service: %s", err)
		return
	}

	vault.SetToken(os.Getenv("VAULT_TOKEN"))
	postgreConf, err := vault.Logical().Read("warscript-users/postgres")
	if err != nil || postgreConf == nil || len(postgreConf.Warnings) != 0 {
		logger.Errorf("can read warscript-users/postges key: %+v; %+v", err, postgreConf)
		return
	}
	redisConf, err := vault.Logical().Read("warscript-users/redis")
	if err != nil || redisConf == nil || len(redisConf.Warnings) != 0 {
		logger.Errorf("can read config/redis key: %s; %+v", err, redisConf.Warnings)
		return
	}

	httpServiceID := fmt.Sprintf("warscript-users-http:%d", httpPort)
	err = consul.Agent().ServiceRegister(&consulapi.AgentServiceRegistration{
		ID:      httpServiceID,
		Name:    "warscript-users-http",
		Port:    httpPort,
		Address: "127.0.0.1",
	})
	defer func() {
		err = consul.Agent().ServiceDeregister(httpServiceID)
		if err != nil {
			logger.Errorf("can not derigister http service: %s", err)
		}
		logger.Info("successfully derigister http service")
	}()

	grpcServiceID := fmt.Sprintf("warscript-users-grpc:%d", grpcPort)
	err = consul.Agent().ServiceRegister(&consulapi.AgentServiceRegistration{
		ID:      grpcServiceID,
		Name:    "warscript-users-grpc",
		Port:    grpcPort,
		Address: "127.0.0.1",
	})
	defer func() {
		err = consul.Agent().ServiceDeregister(grpcServiceID)
		if err != nil {
			logger.Errorf("can not derigister grpc service: %s", err)
		}
		logger.Info("successfully derigister grpc service")
	}()

	rediCli, err = redis.Connect(redisConf.Data["user"].(string),
		redisConf.Data["pass"].(string), redisConf.Data["addr"].(string),
		redisConf.Data["database"].(string))
	if err != nil {
		logger.Errorf("can not connect redis: %s", err)
		return
	}
	defer rediCli.Close()

	pqConn, err = postgresql.Connect(postgreConf.Data["user"].(string), postgreConf.Data["pass"].(string),
		postgreConf.Data["host"].(string), postgreConf.Data["port"].(string), postgreConf.Data["database"].(string))
	if err != nil {
		logger.Errorf("can not connect to postgresql database: %s", err.Error())
		return
	}
	defer pqConn.Close()

	auth := &AuthManager{}
	listenGRPCPort, err := net.Listen("tcp", ":"+strconv.Itoa(grpcPort))
	if err != nil {
		logger.Errorf("grpc port listener error: %s", err)
		return
	}

	serverGRPCAuth := grpc.NewServer()
	models.RegisterAuthServer(serverGRPCAuth, auth)
	logger.Infof("Auth gRPC service successfully started at port %d", grpcPort)
	go func() {
		if err = serverGRPCAuth.Serve(listenGRPCPort); err != nil {
			logger.Fatalf("Auth gRPC service failed at port %d", grpcPort)
			os.Exit(1)
		}
	}()

	localGRPCAuth := &LocalAuthClient{}
	r := mux.NewRouter().PathPrefix("/v1").Subrouter()

	r.HandleFunc("/sessions", middlewares.WithAuthentication(GetSession, logger, localGRPCAuth)).Methods("GET")
	r.HandleFunc("/sessions", CreateSession).Methods("POST")
	r.HandleFunc("/sessions", middlewares.WithAuthentication(DeleteSession, logger, localGRPCAuth)).Methods("DELETE")

	r.HandleFunc("/users", CreateUser).Methods("POST")
	r.HandleFunc("/users", middlewares.WithAuthentication(UpdateUser, logger, localGRPCAuth)).Methods("PUT")
	r.HandleFunc("/users/{user_id:[0-9]+}", GetUser).Methods("GET")
	r.HandleFunc("/users/used", middlewares.WithLimiter(CheckUsername, rate.NewLimiter(3, 5), logger)).Methods("POST")

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/", middlewares.RecoverMiddleware(middlewares.AccessLogMiddleware(r, logger), logger))

	logger.Infof("Auth HTTP service successfully started at port %d", httpPort)
	err = http.ListenAndServe(":"+strconv.Itoa(httpPort), nil)
	if err != nil {
		logger.Errorf("cant start main server. err: %s", err.Error())
		return
	}
}
