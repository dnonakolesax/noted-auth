package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"net"
	"net/http"


	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
	fasthttpprom "github.com/carousell/fasthttp-prometheus-middleware"

	"github.com/dnonakolesax/noted-auth/internal/configs"
	dbredis "github.com/dnonakolesax/noted-auth/internal/db/redis"
	dbsql "github.com/dnonakolesax/noted-auth/internal/db/sql"
	"github.com/dnonakolesax/noted-auth/internal/routing"

	stateRepo "github.com/dnonakolesax/noted-auth/internal/repo/state"
	userRepo "github.com/dnonakolesax/noted-auth/internal/repo/user"

	"github.com/dnonakolesax/noted-auth/internal/usecase"

	authDelivery "github.com/dnonakolesax/noted-auth/internal/delivery/auth/v1"
	userDelivery "github.com/dnonakolesax/noted-auth/internal/delivery/user/v1"
	
	userProto "github.com/dnonakolesax/noted-auth/internal/delivery/user/v1/proto"
	
	"google.golang.org/grpc"

	"github.com/prometheus/client_golang/prometheus/collectors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	/************************************************/
	/*               CONFIG LOADING                 */
	/************************************************/
	v := viper.New()

	err := configs.Load("../../configs/", v)

	if err != nil {
		slog.Error(fmt.Sprintf("Error loading config: %v", err))
		return
	}

	kcConfig := configs.NewKeycloakConfig(v)
	psqlConfig := configs.NewRDBConfig(v)
	redisConfig := configs.NewRedisConfig(v)
	appConfig := configs.NewServiceConfig(v)

	/************************************************/
	/*               SQL DB CONNECTION              */
	/************************************************/

	psqlConn, err := dbsql.NewPGXConn(psqlConfig)
	
	if err != nil {
		slog.Error(fmt.Sprintf("Error connecting to database: %v", err))
		return
	}

	psqlWorker, err := dbsql.NewPGXWorker(psqlConn)

	if err != nil {
		slog.Error(fmt.Sprintf("Error creating worker: %v", err))
		return
	}
	
	/************************************************/
	/*              REDIS DB CONNECTION             */
	/************************************************/

	redisClient := dbredis.NewClient(redisConfig)

	/************************************************/
	/*                METRICS SETUP                 */
	/************************************************/
	
	reg := prometheus.NewRegistry()

	kcRequestDurations := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "keycloak_post_request_duration_ms",
		Help:    "A histogram of the keycloak token POST request durations in ms.",
		Buckets: prometheus.ExponentialBuckets(0.1, 1.5, 5),
	})

	kcRequestOks := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "keycloak_post_request_200s",
		Help: "The total number of 200 keycloak token POST requests.",
	})

	kcRequest400s := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "keycloak_post_request_400s",
		Help: "The total number of 400 keycloak token POST requests.",
	})

	kcRequest500s := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "keycloak_post_request_500s",
		Help: "The total number of 500 keycloak token POST requests.",
	})

	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		kcRequestDurations,
		kcRequestOks,
		kcRequest400s,
		kcRequest500s,
	)

	tokenMetrics := usecase.TokenMetrics{
		RequestDurations: kcRequestDurations,
		RequestOks: kcRequestOks,
		RequestBads: kcRequest400s,
		RequestServErrs: kcRequest500s,
	}

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	
	go func() {
		err := http.ListenAndServe(":"+string(appConfig.MetricsPort), nil)
		if err != nil {
			slog.Error(fmt.Sprintf("Error starting metrics server: %v", err))
			return 
		}
	}()
	/************************************************/
	/*                  REPOS INIT                  */
	/************************************************/

	stateRepository := stateRepo.NewRedisStateRepo(redisClient)
	userRepository, err := userRepo.NewUserRepo(psqlWorker)

	if err != nil {
		slog.Error(fmt.Sprintf("Error creating user repository: %v", err))
		return
	}

	/************************************************/
	/*                USECASES INIT                 */
	/************************************************/

	stateUsecase := usecase.NewAuthUsecase(appConfig.AuthTimeout, stateRepository, kcConfig, tokenMetrics)
	userUsecase := usecase.NewUserUsecase(userRepository)

	/************************************************/
	/*              REST HANDLERS INIT              */
	/************************************************/

	authHandler := authDelivery.NewAuthHandler(appConfig.AllowedRedirect, appConfig.AllowedRedirect, stateUsecase)
	userHandler := userDelivery.NewUserHandler(userUsecase)

	
	/************************************************/
	/*               HTTP ROUTER SETUP              */
	/************************************************/

	router := routing.NewRouter()
	p := fasthttpprom.NewPrometheus("")
	p.Use(router.Router())
	router.NewApiGroup(appConfig.BasePath, "1", authHandler, userHandler)

	/************************************************/
	/*               GRPC SERVER START              */
	/************************************************/
	listen, err := net.Listen("tcp", ":8082")

	if err != nil {
		slog.Error(fmt.Sprintf("Error listening grpc net: %v", err))
		return
	}
	
	grpcSrv := grpc.NewServer()
	userProto.RegisterUserServiceServer(grpcSrv, userDelivery.NewUserService(userUsecase))

	go func() {
		err = grpcSrv.Serve(listen)

		if err != nil {
			slog.Error(fmt.Sprintf("Error starting grpc server: %v", err))
		}
	}()



	/************************************************/
	/*               HTTP SERVER START              */
	/************************************************/

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	srv := fasthttp.Server{
		Handler: router.Router().Handler,
	}

	go func() {
		err := srv.ListenAndServe(":" + string(appConfig.Port), )
		if err != nil {
			slog.Error(fmt.Sprintf("Couldn't start server: %v", err))
		}
	}()
	
	sig := <-quit
	slog.Info(fmt.Sprintf("Received stop signal: %v", sig))
	err = srv.Shutdown()
	if err != nil {
		slog.Error(fmt.Sprintf("shutdown returned err: %s \n", err))
	}
}
