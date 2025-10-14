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
	"sync"
	"sync/atomic"
	"syscall"

	fasthttpprom "github.com/carousell/fasthttp-prometheus-middleware"
	"github.com/dnonakolesax/viper"
	"github.com/valyala/fasthttp"

	"github.com/dnonakolesax/noted-auth/internal/configs"
	"github.com/dnonakolesax/noted-auth/internal/consts"
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
	healthDelivery "github.com/dnonakolesax/noted-auth/internal/delivery/healthcheck/v1"
	sessionDelivery "github.com/dnonakolesax/noted-auth/internal/delivery/session/v1"
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
// @BasePath /api/v1/iam.
func main() { //nolint:funlen // TODO: refactor
	initLogger := logger.NewLogger("info", true, "init")
	/************************************************/
	/*               CONFIG LOADING                 */
	/************************************************/

	vaultHealthcheck := &atomic.Bool{}
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

	err := configs.Load("./configs/", v, initLogger, &kcConfig, &psqlConfig, &redisConfig,
		&appConfig, &serverConfig, &httpClientConfig, &loggerConfig)

	if err != nil {
		initLogger.ErrorContext(context.Background(), "Error loading config",
			slog.String(consts.ErrorLoggerKey, err.Error()))
		return
	}
	/************************************************/
	/*                 LOGGER SETUP                 */
	/************************************************/

	appLoggers := logger.SetupLoggers(loggerConfig.LogLevel, loggerConfig.LogAddSource)

	/************************************************/
	/*               SQL DB CONNECTION              */
	/************************************************/

	postgresHealthcheck := &atomic.Bool{}
	initLogger.InfoContext(context.Background(), "Starting SQL DB connection")
	psqlConn, err := dbsql.NewPGXConn(psqlConfig, appLoggers.Infra)
	initLogger.InfoContext(context.Background(), "SQL DB connection established")

	if err != nil {
		initLogger.ErrorContext(context.Background(), "Error connecting to database",
			slog.String(consts.ErrorLoggerKey, err.Error()))
		return
	}
	defer psqlConn.Disconnect()

	psqlWorker, err := dbsql.NewPGXWorker(psqlConn)

	if err != nil {
		initLogger.ErrorContext(context.Background(), "Error creating pgsql worker",
			slog.String(consts.ErrorLoggerKey, err.Error()))
		return
	}

	/************************************************/
	/*              REDIS DB CONNECTION             */
	/************************************************/

	redisHealtcheck := &atomic.Bool{}
	initLogger.InfoContext(context.Background(), "Starting REDIS DB connection")
	redisClient, err := dbredis.NewClient(redisConfig, appLoggers.Infra)
	initLogger.InfoContext(context.Background(), "REDIS DB connection established")

	if err != nil {
		initLogger.ErrorContext(context.Background(), "Error connecting to redis",
			slog.String(consts.ErrorLoggerKey, err.Error()))
		return
	}
	defer redisClient.Client.Close()

	/************************************************/
	/*               SHUTDOWN WG SETUP              */
	/************************************************/

	wg := &sync.WaitGroup{}

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
		Handler:           promhttp.Handler(),
		Addr:              ":" + strconv.Itoa(appConfig.MetricsPort),
		ReadHeaderTimeout: serverConfig.ReadTimeout,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("Starting metrics server", slog.Int("Port", appConfig.MetricsPort))
		msErr := metricsServer.ListenAndServe()
		if msErr != nil && msErr != http.ErrServerClosed {
			slog.Error(fmt.Sprintf("Error starting metrics server: %v", err))
			panic(err)
		}
	}()

	/************************************************/
	/*              HTTP CLIENT SETUP               */
	/************************************************/

	keycloakHealthcheck := &atomic.Bool{}
	httpClient := httpclient.NewWithRetry(kcConfig.InterRealmAddress+kcConfig.TokenEndpoint,
		httpClientConfig, tokenRequestMetrics, appLoggers.HTTPc)

	/************************************************/
	/*                  REPOS INIT                  */
	/************************************************/

	stateRedisRepository := stateRepo.NewRedisStateRepo(redisClient, appLoggers.Repo)
	stateInMemoryRepository := stateRepo.NewInMemStateRepo(appLoggers.Repo)
	stateRepos := []usecase.StateRepo{stateInMemoryRepository, stateRedisRepository}
	userRepository, err := userRepo.NewUserRepo(psqlWorker, kcConfig.RealmID, psqlConfig.RequestsPath, appLoggers.Repo)

	if err != nil {
		initLogger.ErrorContext(context.Background(), "Error creating user repository",
			slog.String(consts.ErrorLoggerKey, err.Error()))
		panic(fmt.Errorf("error creating user repository %s", err.Error()))
	}

	/************************************************/
	/*                USECASES INIT                 */
	/************************************************/

	stateUsecase := usecase.NewAuthUsecase(appConfig.AuthTimeout, stateRepos, kcConfig, httpClient,
		appLoggers.Service)
	userUsecase := usecase.NewUserUsecase(userRepository, appLoggers.Service)
	sessionUsecase := usecase.NewSessionUsecase(httpClient, appLoggers.Service)

	/************************************************/
	/*              REST HANDLERS INIT              */
	/************************************************/

	authHandler := authDelivery.NewAuthHandler(appConfig.AllowedRedirect, appConfig.AllowedRedirect,
		stateUsecase, appLoggers.HTTP)
	userHandler := userDelivery.NewUserHandler(userUsecase, appLoggers.HTTP)
	sessionHandler := sessionDelivery.NewSessionHandler(sessionUsecase, appLoggers.HTTP)
	healthcheckHandler := healthDelivery.NewHealthCheckHandler(redisHealtcheck, postgresHealthcheck,
		keycloakHealthcheck, vaultHealthcheck, appLoggers.HTTP)

	/************************************************/
	/*               HTTP ROUTER SETUP              */
	/************************************************/

	router := routing.NewRouter()
	p := fasthttpprom.NewPrometheus("")
	p.Use(router.Router())
	router.NewAPIGroup(appConfig.BasePath, "1", authHandler, userHandler, sessionHandler, healthcheckHandler)

	/************************************************/
	/*               GRPC SERVER START              */
	/************************************************/

	cfg := net.ListenConfig{}
	listener, err := cfg.Listen(context.Background(), "tcp", ":"+strconv.Itoa(appConfig.GRPCPort))

	if err != nil {
		initLogger.ErrorContext(context.Background(), "Error listening grpc net",
			slog.String(consts.ErrorLoggerKey, err.Error()))
		panic(fmt.Sprintf("error listening grpc net: %v", err))
	}

	grpcSrv := grpc.NewServer()
	userProto.RegisterUserServiceServer(grpcSrv, userDelivery.NewUserServer(userUsecase, appLoggers.GRPC))

	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("Starting GRPC server", slog.Int("Port", appConfig.GRPCPort))
		err = grpcSrv.Serve(listener)

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

	wg.Add(1)
	go func() {
		defer wg.Done()
		slog.Info("Starting HTTP server", slog.Int("Port", appConfig.Port))
		httpErr := srv.ListenAndServe(":" + strconv.Itoa(appConfig.Port))
		if httpErr != nil {
			slog.Error(fmt.Sprintf("Couldn't start server: %v", err))
		}
	}()

	/************************************************/
	/*               HTTP SERVER STOP               */
	/************************************************/
	sig := <-quit
	initLogger.InfoContext(context.Background(), "Received signal", slog.String("signal", sig.String()))
	err = srv.Shutdown()
	if err != nil {
		initLogger.ErrorContext(context.Background(), "Main HTTP server shutdown error",
			slog.String(consts.ErrorLoggerKey, err.Error()))
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
		initLogger.ErrorContext(context.Background(), "Metrics server shutdown error",
			slog.String(consts.ErrorLoggerKey, err.Error()))
	}

	wg.Wait()
}
