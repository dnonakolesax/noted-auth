package application

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/dnonakolesax/noted-auth/internal/consts"

	stateRepo "github.com/dnonakolesax/noted-auth/internal/repo/state"
	userRepo "github.com/dnonakolesax/noted-auth/internal/repo/user"

	"github.com/dnonakolesax/noted-auth/internal/usecase"

	authDelivery "github.com/dnonakolesax/noted-auth/internal/delivery/auth/v1"
	healthDelivery "github.com/dnonakolesax/noted-auth/internal/delivery/healthcheck/v1"
	sessionDelivery "github.com/dnonakolesax/noted-auth/internal/delivery/session/v1"
	userDelivery "github.com/dnonakolesax/noted-auth/internal/delivery/user/v1"
)

type Layers struct {
	authHTTP    *authDelivery.Handler
	hcHTTP      *healthDelivery.Handler
	sessionHTTP *sessionDelivery.Handler
	userHTTP    *userDelivery.Handler
	userGRPC    *userDelivery.Server

	// authUsecase    usecase.AuthUsecase
	// sessionUsecase usecase.SessionUsecase
	// userUsecase    usecase.UserUsecase

	// stateRepoRedis stateRepo.RedisStateRepo
	// stateRepoInMem stateRepo.InMemStateRepo
	// userRepo       userRepo.UserRepo
}

func (a *App) SetupLayers() (error) {
	/************************************************/
	/*                  REPOS INIT                  */
	/************************************************/

	stateRedisRepository := stateRepo.NewRedisStateRepo(a.components.redis, a.loggers.Repo)
	stateInMemoryRepository := stateRepo.NewInMemStateRepo(a.loggers.Repo)
	stateRepos := []usecase.StateRepo{stateInMemoryRepository, stateRedisRepository}
	userRepository, err := userRepo.NewUserRepo(a.components.pgsql, a.configs.Keycloak.RealmID,
		 a.configs.PSQL.RequestsPath, a.loggers.Repo)

	if err != nil {
		a.initLogger.ErrorContext(context.Background(), "Error creating user repository",
			slog.String(consts.ErrorLoggerKey, err.Error()))
		return fmt.Errorf("error creating user repository %s", err.Error())
	}

	/************************************************/
	/*                USECASES INIT                 */
	/************************************************/

	stateUsecase := usecase.NewAuthUsecase(a.configs.Service.AuthTimeout, stateRepos, *a.configs.Keycloak, a.components.keycloak,
		a.loggers.Service, a.configs.UpdateChans.KCClientSecret)
	userUsecase := usecase.NewUserUsecase(userRepository, a.loggers.Service)
	sessionUsecase := usecase.NewSessionUsecase(a.components.keycloak2, a.loggers.Service)

	/************************************************/
	/*              REST HANDLERS INIT              */
	/************************************************/

	authHandler := authDelivery.NewAuthHandler(a.configs.Service.AllowedRedirect, a.configs.Service.AllowedRedirect,
		stateUsecase, a.loggers.HTTP)
	userHandler := userDelivery.NewUserHandler(userUsecase, a.loggers.HTTP, stateUsecase)
	sessionHandler := sessionDelivery.NewSessionHandler(sessionUsecase, a.loggers.HTTP)
	healthcheckHandler := healthDelivery.NewHealthCheckHandler(a.health.Redis, a.health.Postgres,
		 a.health.Keycloak, a.health.Vault, a.loggers.HTTP)

	userServer := userDelivery.NewUserServer(userUsecase, a.loggers.GRPC)

	a.layers = &Layers{
		authHTTP: authHandler,
		userHTTP: userHandler,
		sessionHTTP: sessionHandler,
		userGRPC: userServer,
		hcHTTP: healthcheckHandler,
	}
	return nil
}
