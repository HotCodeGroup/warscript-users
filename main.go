package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"github.com/HotCodeGroup/warscript-utils/logging"
	"github.com/HotCodeGroup/warscript-utils/middlewares"
	"github.com/HotCodeGroup/warscript-utils/models"
	"github.com/HotCodeGroup/warscript-utils/postgresql"
	"github.com/HotCodeGroup/warscript-utils/redis"

	"google.golang.org/grpc"
)

var logger *logrus.Logger

func main() {
	var err error
	logger, err = logging.NewLogger(os.Stdout, os.Getenv("LOGENTRIESRUS_TOKEN"))
	if err != nil {
		log.Printf("can not create logger: %s", err)
		return
	}

	rediCli, err = redis.Connect(os.Getenv("STORAGE_USER"), os.Getenv("STORAGE_PASS"), os.Getenv("STORAGE_HOST"))
	if err != nil {
		logger.Errorf("can not connect redis: %s", err)
		return
	}
	defer rediCli.Close()

	pgxConn, err = postgresql.Connect(os.Getenv("DB_USER"), os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	if err != nil {
		logger.Errorf("can not connect to postgresql database: %s", err.Error())
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
		if err = serverGRPCAuth.Serve(listenGRPCPort); err != nil {
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
