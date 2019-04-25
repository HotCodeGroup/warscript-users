package main

import (
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx"

	"github.com/go-redis/redis"
	"github.com/jcftang/logentriesrus"

	log "github.com/sirupsen/logrus"
)

type AuthManager struct {
	RedisCli *redis.Client
	PgCli    *pgx.ConnPool

	Router http.Handler
}

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	// собираем логи в хранилище
	le, err := logentriesrus.NewLogentriesrusHook(os.Getenv("LOGENTRIESRUS_TOKEN"))
	if err != nil {
		os.Exit(-1)
	}
	log.AddHook(le)
}

func main() {
	dbPort, err := strconv.ParseInt(os.Getenv("DB_PORT"), 10, 16)
	if err != nil {
		log.Errorf("incorrect database port: %s", err.Error())
		return
	}

	rCli := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("STORAGE_HOST"),
		Password: os.Getenv("STORAGE_PASS"),
		DB:       0,
	})
	_, err = rCli.Ping().Result()
	if err != nil {
		log.Errorf("cant connect to resis storage: %s", err.Error())
		return
	}
	defer rCli.Close()

	pgCli, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     os.Getenv("DB_HOST"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASS"),
			Database: os.Getenv("DB_NAME"),
			Port:     uint16(dbPort),
		},
	})
	if err != nil {
		log.Errorf("cant connect to postgresql database: %s", err.Error())
		return
	}
	defer pgCli.Close()

	r := mux.NewRouter().PathPrefix("/v1").Subrouter()

	auth := &AuthManager{
		RedisCli: rCli,
		PgCli:    pgCli,
		Router:   r,
	}

	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins([]string{os.Getenv("CORS_HOST")}),
		handlers.AllowedMethods([]string{"POST", "GET", "PUT", "DELETE"}),
		handlers.AllowedHeaders([]string{"Content-Type"}),
		handlers.AllowCredentials(),
	)

	port := os.Getenv("PORT")
	log.Printf("MainService successfully started at port %s", port)
	err = http.ListenAndServe(":"+port, corsMiddleware(auth.Router))
	if err != nil {
		log.Errorf("cant start main server. err: %s", err.Error())
		return
	}
}
