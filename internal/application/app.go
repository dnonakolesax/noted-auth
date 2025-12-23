package application

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
	"syscall"

	fasthttpprom "github.com/carousell/fasthttp-prometheus-middleware"
	"github.com/dnonakolesax/noted-auth/internal/configs"
	"github.com/dnonakolesax/noted-auth/internal/consts"
	"github.com/dnonakolesax/noted-auth/internal/logger"
	"github.com/dnonakolesax/noted-auth/internal/middlewares"
	"github.com/dnonakolesax/noted-auth/internal/routing"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp"

	userProto "github.com/dnonakolesax/noted-auth/internal/delivery/user/v1/proto"
	authProto "github.com/dnonakolesax/noted-auth/internal/delivery/auth/v1/proto"
	"google.golang.org/grpc"
)

type App struct {
	configs    *configs.Config
	health     *HealthChecks
	metrics    *Metrics
	initLogger *slog.Logger
	layers     *Layers
	loggers    *logger.Loggers
	components *Components
}

func NewApp(configsDir string) (*App, error) {
	lcfg := &configs.LoggerConfig{LogLevel: "info", LogAddSource: true}
	initLogger := logger.NewLogger(lcfg, "init")
	app := &App{}

	app.initLogger = initLogger

	app.InitHealthchecks()

	configs, err := configs.SetupConfigs(initLogger, configsDir, app.health.Vault)

	if err != nil {
		return nil, err
	}

	app.configs = configs

	loggers := logger.SetupLoggers(app.configs.Logger)

	app.loggers = loggers

	app.SetupMetrics()

	err = app.SetupComponents()

	if err != nil {
		return nil, err
	}

	err = app.SetupLayers()

	if err != nil {
		return nil, err
	}

	return app, nil
}

func (a *App) Run() {
	/************************************************/
	/*               HTTP ROUTER SETUP              */
	/************************************************/

	router := routing.NewRouter()
	p := fasthttpprom.NewPrometheus("")
	p.Use(router.Router())
	router.NewAPIGroup(a.configs.Service.BasePath, "1",
		a.layers.authHTTP, a.layers.userHTTP, a.layers.sessionHTTP, a.layers.hcHTTP)

	wg := &sync.WaitGroup{}

	/************************************************/
	/*               GRPC SERVER START              */
	/************************************************/

	cfg := net.ListenConfig{}
	listener, err := cfg.Listen(context.Background(), "tcp", ":"+strconv.Itoa(
		a.configs.Service.GRPCPort,
	))

	if err != nil {
		a.initLogger.ErrorContext(context.Background(), "Error listening grpc net",
			slog.String(consts.ErrorLoggerKey, err.Error()))
		panic(fmt.Sprintf("error listening grpc net: %v", err))
	}

	grpcSrv := grpc.NewServer()
	userProto.RegisterUserServiceServer(grpcSrv, a.layers.userGRPC)
	authProto.RegisterAuthServiceServer(grpcSrv, a.layers.authGRPC)

	wg.Add(1)
	go func() {
		defer wg.Done()
		a.initLogger.Info("Starting GRPC server", slog.Int("Port", a.configs.Service.GRPCPort))
		err = grpcSrv.Serve(listener)

		if err != nil {
			a.initLogger.Error(fmt.Sprintf("Error starting grpc server: %v", err))
			panic(err)
		}
	}()

	/************************************************/
	/*               HTTP SERVER START              */
	/************************************************/

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	srv := fasthttp.Server{
		Handler: middlewares.CommonMiddleware(router.Router().Handler, a.loggers.HTTP),

		ReadTimeout:  a.configs.HTTPServer.ReadTimeout,
		WriteTimeout: a.configs.HTTPServer.WriteTimeout,
		IdleTimeout:  a.configs.HTTPServer.IdleTimeout,

		MaxRequestBodySize: a.configs.HTTPServer.MaxReqBodySize,
		ReadBufferSize:     a.configs.HTTPServer.ReadBufferSize,
		WriteBufferSize:    a.configs.HTTPServer.WriteBufferSize,

		Concurrency:        a.configs.HTTPServer.Concurrency,
		MaxConnsPerIP:      a.configs.HTTPServer.MaxConnsPerIP,
		MaxRequestsPerConn: a.configs.HTTPServer.MaxRequestsPerConn,

		TCPKeepalivePeriod: a.configs.HTTPServer.TCPKeepAlivePeriod,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		a.initLogger.Info("Starting HTTP server", slog.Int("Port", a.configs.Service.Port))
		httpErr := srv.ListenAndServe(":" + strconv.Itoa(a.configs.Service.Port))
		if httpErr != nil {
			a.initLogger.Error(fmt.Sprintf("Couldn't start server: %v", err))
		}
	}()

	/************************************************/
	/*             METRICS SERVER START             */
	/************************************************/
	metricsServer := &http.Server{
		Handler:           promhttp.HandlerFor(a.metrics.Reg, promhttp.HandlerOpts{Registry: a.metrics.Reg}),
		Addr:              ":" + strconv.Itoa(a.configs.Service.MetricsPort),
		ReadHeaderTimeout: a.configs.HTTPServer.ReadTimeout,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		a.initLogger.Info("Starting metrics server", slog.Int("Port", a.configs.Service.MetricsPort))
		msErr := metricsServer.ListenAndServe()
		if msErr != nil && msErr != http.ErrServerClosed {
			a.initLogger.Error(fmt.Sprintf("Error starting metrics server: %v", err))
			panic(err)
		}
	}()

	/************************************************/
	/*               HTTP SERVER STOP               */
	/************************************************/
	sig := <-quit
	a.initLogger.InfoContext(context.Background(), "Received signal", slog.String("signal", sig.String()))
	err = srv.Shutdown()
	if err != nil {
		a.initLogger.ErrorContext(context.Background(), "Main HTTP server shutdown error",
			slog.String(consts.ErrorLoggerKey, err.Error()))
	}

	/************************************************/
	/*               GRPC SERVER STOP               */
	/************************************************/

	grpcSrv.Stop()

	/************************************************/
	/*             METRICS SERVER STOP              */
	/************************************************/

	ctx, cancel := context.WithTimeout(context.Background(), a.configs.HTTPServer.IdleTimeout)
	defer cancel()
	err = metricsServer.Shutdown(ctx)

	if err != nil {
		a.initLogger.ErrorContext(context.Background(), "Metrics server shutdown error",
			slog.String(consts.ErrorLoggerKey, err.Error()))
	}

	a.components.pgsql.Conn.Disconnect()
	_ = a.components.redis.Client.Close()

	wg.Wait()
}
