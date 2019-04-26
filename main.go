package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"golang.org/x/time/rate"

	"github.com/HotCodeGroup/warscript-utils/middlewares"
	"github.com/HotCodeGroup/warscript-utils/models"
	"github.com/jackc/pgx"
	"google.golang.org/grpc"

	"github.com/go-redis/redis"
	"github.com/jcftang/logentriesrus"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func main() {
	logger = logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)

	// собираем логи в хранилище
	le, err := logentriesrus.NewLogentriesrusHook(os.Getenv("LOGENTRIESRUS_TOKEN"))
	if err != nil {
		log.Printf("can not create logrus logger %s", err)
		return
	}
	logger.AddHook(le)

	dbPort, err := strconv.ParseInt(os.Getenv("DB_PORT"), 10, 16)
	if err != nil {
		logger.Errorf("incorrect database port: %s", err.Error())
		return
	}

	rediCli = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("STORAGE_HOST"),
		Password: os.Getenv("STORAGE_PASS"),
		DB:       0,
	})
	_, err = rediCli.Ping().Result()
	if err != nil {
		logger.Errorf("cant connect to resis storage: %s", err.Error())
		return
	}
	defer rediCli.Close()

	pgxConn, err = pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     os.Getenv("DB_HOST"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASS"),
			Database: os.Getenv("DB_NAME"),
			Port:     uint16(dbPort),
		},
	})
	if err != nil {
		logger.Errorf("cant connect to postgresql database: %s", err.Error())
		return
	}
	defer pgxConn.Close()

	auth := &AuthManager{}
	grpcPort := os.Getenv("GRPC_PORT")
	listenGRPCPort, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		logger.Errorf("grpc port listener error: %s", err)
		return
	}

	serverGRPCAuth := grpc.NewServer()
	models.RegisterAuthServer(serverGRPCAuth, auth)

	logger.Infof("Auth gRPC service successfully started at port %s", grpcPort)
	go func() {
		if err := serverGRPCAuth.Serve(listenGRPCPort); err != nil {
			logger.Fatalf("Auth gRPC service failed at port %s", grpcPort)
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

	httpPort := os.Getenv("HTTP_PORT")
	logger.Infof("Auth HTTP service successfully started at port %s", httpPort)
	err = http.ListenAndServe(":"+httpPort, r)
	if err != nil {
		logger.Errorf("cant start main server. err: %s", err.Error())
		return
	}
}
