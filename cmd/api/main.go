package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	fasthttpprom "github.com/carousell/fasthttp-prometheus-middleware"
	"github.com/dnonakolesax/viper"
	"github.com/valyala/fasthttp"

	"github.com/dnonakolesax/noted-auth/internal/configs"
	dbredis "github.com/dnonakolesax/noted-auth/internal/db/redis"
	dbsql "github.com/dnonakolesax/noted-auth/internal/db/sql"
	"github.com/dnonakolesax/noted-auth/internal/httpclient"
	"github.com/dnonakolesax/noted-auth/internal/logger"
	"github.com/dnonakolesax/noted-auth/internal/metrics"
	"github.com/dnonakolesax/noted-auth/internal/middlewares"
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

// @title OIDC API
// @version 1.0
// @description API for authorising users and storing their info

// @contact.name G
// @contact.email bg@dnk33.com

// @host oauth.dnk33.com
// @BasePath /api/v1/iam
func main() {
	/************************************************/
	/*                PANIC CATCHER                 */
	/************************************************/

	defer func() {
		err := recover()

		if err != nil {
			slog.Error("Panic caught in main: ", slog.Any("Error", err))
		}
	}()

	/************************************************/
	/*               CONFIG LOADING                 */
	/************************************************/

	v := viper.New()
	v.PanicOnNil = true
	// ondefault

	kcConfig := configs.KeycloakConfig{}
	psqlConfig := configs.RDBConfig{}
	redisConfig := configs.RedisConfig{}
	appConfig := configs.ServiceConfig{}
	serverConfig := configs.HTTPServerConfig{}
	httpClientConfig := configs.HTTPClientConfig{}
	loggerConfig := configs.LoggerConfig{}

	err := configs.Load("./configs/", v, &kcConfig, &psqlConfig, &redisConfig, &appConfig, &serverConfig, &httpClientConfig, &loggerConfig)

	if err != nil {
		slog.Error(fmt.Sprintf("Error loading config: %v", err))
		return
	}
	/************************************************/
	/*                 LOGGER SETUP                 */
	/************************************************/

	appLogger := logger.NewLogger(loggerConfig.LogLevel, loggerConfig.LogAddSource)
	slog.SetDefault(appLogger)

	/************************************************/
	/*               SQL DB CONNECTION              */
	/************************************************/

	psqlConn, err := dbsql.NewPGXConn(psqlConfig)

	if err != nil {
		slog.Error(fmt.Sprintf("Error connecting to database: %v", err))
		return
	}
	defer psqlConn.Disconnect()

	psqlWorker, err := dbsql.NewPGXWorker(psqlConn)

	if err != nil {
		slog.Error(fmt.Sprintf("Error creating worker: %v", err))
		return
	}

	/************************************************/
	/*              REDIS DB CONNECTION             */
	/************************************************/

	redisClient, err := dbredis.NewClient(redisConfig)

	if err != nil {
		slog.Error(fmt.Sprintf("Error connecting to redis: %v", err))
		return
	}
	defer redisClient.Client.Close()

	/************************************************/
	/*                METRICS SETUP                 */
	/************************************************/

	reg := prometheus.NewRegistry()

	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	tokenRequestMetrics := metrics.NewHTTPRequestMetrics(reg, "keycloak_token_post")

	metricsServer := &http.Server{
		Handler: promhttp.Handler(),
		Addr:    ":" + strconv.Itoa(int(appConfig.MetricsPort)),
	}

	go func() {
		slog.Info("Starting metrics server", slog.Int("Port", int(appConfig.MetricsPort)))
		err := metricsServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			slog.Error(fmt.Sprintf("Error starting metrics server: %v", err))
			panic(err)
		}
	}()

	/************************************************/
	/*              HTTP CLIENT SETUP               */
	/************************************************/

	httpClient := httpclient.NewWithRetry(kcConfig.InterRealmAddress+kcConfig.TokenEndpoint, httpClientConfig, tokenRequestMetrics)

	/************************************************/
	/*                  REPOS INIT                  */
	/************************************************/

	stateRepository := stateRepo.NewRedisStateRepo(redisClient)
	userRepository, err := userRepo.NewUserRepo(psqlWorker, kcConfig.RealmId)

	if err != nil {
		slog.Error(fmt.Sprintf("Error creating user repository: %v", err))
		return
	}

	/************************************************/
	/*                USECASES INIT                 */
	/************************************************/

	stateUsecase := usecase.NewAuthUsecase(appConfig.AuthTimeout, stateRepository, kcConfig, httpClient)
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

	listen, err := net.Listen("tcp", ":"+strconv.Itoa(int(appConfig.GRPCPort)))

	if err != nil {
		slog.Error(fmt.Sprintf("Error listening grpc net: %v", err))
		panic(err)
	}

	grpcSrv := grpc.NewServer()
	userProto.RegisterUserServiceServer(grpcSrv, userDelivery.NewUserService(userUsecase))

	go func() {
		slog.Info("Starting GRPC server", slog.Int("Port", int(appConfig.GRPCPort)))
		err = grpcSrv.Serve(listen)

		if err != nil {
			slog.Error(fmt.Sprintf("Error starting grpc server: %v", err))
			panic(err)
		}
	}()

	/************************************************/
	/*               HTTP SERVER START              */
	/************************************************/

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	srv := fasthttp.Server{
		Handler: middlewares.CommonMiddleware(router.Router().Handler),

		ReadTimeout:  serverConfig.ReadTimeout,
		WriteTimeout: serverConfig.WriteTimeout,
		IdleTimeout:  serverConfig.IdleTimeout,

		MaxRequestBodySize: serverConfig.MaxReqBodySize,
		ReadBufferSize:     serverConfig.ReadBufferSize,
		WriteBufferSize:    serverConfig.WriteBufferSize,

		Concurrency:        serverConfig.Concurrency,
		MaxConnsPerIP:      serverConfig.MaxConnsPerIP,
		MaxRequestsPerConn: serverConfig.MaxRequestsPerConn,

		TCPKeepalivePeriod: serverConfig.TCPKeepAlivePeriod,
	}

	go func() {
		slog.Info("Starting HTTP server", slog.Int("Port", int(appConfig.Port)))
		err := srv.ListenAndServe(":" + strconv.Itoa(int(appConfig.Port)))
		if err != nil {
			slog.Error(fmt.Sprintf("Couldn't start server: %v", err))
		}
	}()

	/************************************************/
	/*               HTTP SERVER STOP               */
	/************************************************/
	sig := <-quit
	slog.Info(fmt.Sprintf("Received stop signal: %v", sig))
	err = srv.Shutdown()
	if err != nil {
		slog.Error(fmt.Sprintf("Main HTTP shutdown returned err: %s \n", err))
	}

	/************************************************/
	/*               GRPC SERVER STOP               */
	/************************************************/

	grpcSrv.Stop()

	/************************************************/
	/*             METRICS SERVER STOP              */
	/************************************************/

	ctx, cancel := context.WithTimeout(context.Background(), serverConfig.IdleTimeout)
	defer cancel()
	err = metricsServer.Shutdown(ctx)

	if err != nil {
		slog.Error(fmt.Sprintf("Metrics HTTP shutdown returned err: %s \n", err))
	}
}
