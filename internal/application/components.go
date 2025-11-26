package application

import (
	"context"
	"log/slog"

	"github.com/dnonakolesax/noted-auth/internal/consts"
	dbredis "github.com/dnonakolesax/noted-auth/internal/db/redis"
	dbsql "github.com/dnonakolesax/noted-auth/internal/db/sql"
	"github.com/dnonakolesax/noted-auth/internal/httpclient"
)

type Components struct {
	redis    *dbredis.Client
	pgsql    *dbsql.PGXWorker
	keycloak *httpclient.HTTPClient
	keycloak2 *httpclient.HTTPClient
}

func (a *App) SetupComponents() (error) {
	/************************************************/
	/*               SQL DB CONNECTION              */
	/************************************************/
	a.initLogger.InfoContext(context.Background(), "Starting SQL DB connection")
	psqlConn, err := dbsql.NewPGXConn(*a.configs.PSQL, a.loggers.Infra)
	a.initLogger.InfoContext(context.Background(), "SQL DB connection established")

	if err != nil {
		a.initLogger.ErrorContext(context.Background(), "Error connecting to database",
			slog.String(consts.ErrorLoggerKey, err.Error()))
		return err
	}

	psqlWorker, err := dbsql.NewPGXWorker(psqlConn, a.health.Postgres, a.configs.UpdateChans.PSQLCredentials)

	if err != nil {
		a.initLogger.ErrorContext(context.Background(), "Error creating pgsql worker",
			slog.String(consts.ErrorLoggerKey, err.Error()))
		return err
	}

	/************************************************/
	/*              REDIS DB CONNECTION             */
	/************************************************/
	a.initLogger.InfoContext(context.Background(), "Starting REDIS DB connection")
	redisClient, err := dbredis.NewClient(a.configs.Redis, a.health.Redis, a.loggers.Infra, a.configs.UpdateChans.RedisPassword)
	a.initLogger.InfoContext(context.Background(), "REDIS DB connection established")

	if err != nil {
		a.initLogger.ErrorContext(context.Background(), "Error connecting to redis",
			slog.String(consts.ErrorLoggerKey, err.Error()))
		return err
	}

	/************************************************/
	/*              HTTP CLIENT SETUP               */
	/************************************************/

	a.initLogger.InfoContext(context.Background(), "Creating HTTP client")
	httpClient, err := httpclient.NewWithRetry(a.configs.Keycloak.InterRealmAddress+a.configs.Keycloak.TokenEndpoint,
		a.configs.HTTPClient, a.metrics.TokenGetMetrics, a.health.Keycloak, a.loggers.HTTPc)

	if err != nil {
		a.initLogger.ErrorContext(context.Background(), "Error connecting to keycloak",
			slog.String(consts.ErrorLoggerKey, err.Error()))
		return err
	}	

	a.initLogger.InfoContext(context.Background(), "Created HTTP client, keycloak pinged")
	a.components =  &Components{
		pgsql: psqlWorker,
		redis: redisClient,
		keycloak: httpClient,
	}

	/************************************************/
	/*              HTTP CLIENT SETUP               */
	/************************************************/

	a.initLogger.InfoContext(context.Background(), "Creating HTTP client for sessions")
	httpClient2, err := httpclient.NewWithRetry("http://eager_faraday:8080/realms/noted/account/sessions/devices/",
		a.configs.HTTPClient, a.metrics.TokenGetMetrics, a.health.Keycloak, a.loggers.HTTPc)

	if err != nil {
		a.initLogger.ErrorContext(context.Background(), "Error connecting to keycloak",
			slog.String(consts.ErrorLoggerKey, err.Error()))
		return err
	}	

	a.initLogger.InfoContext(context.Background(), "Created HTTP client, keycloak pinged")
	a.components =  &Components{
		pgsql: psqlWorker,
		redis: redisClient,
		keycloak: httpClient,
		keycloak2: httpClient2,
	}
	return nil
}
