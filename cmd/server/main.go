package main

import (
	"context"
	"loan-service/cmd/base"
	memoryDB "loan-service/internal/infrastructure/memory"
	postgresDB "loan-service/internal/infrastructure/postgres"
	"loan-service/internal/interface/http"
	"loan-service/internal/usecase"
	log "loan-service/pkg/logger"
	"loan-service/pkg/memory"
	"loan-service/pkg/postgres"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
)

func main() {
	logger := log.NewLogger()

	var env base.Env

	if _, err := flags.Parse(&env); err != nil {
		if flags.WroteHelp(err) {
			os.Exit(0)
		}
		// * Sleep 1 second
		time.Sleep(time.Second)

		logger.Fatal().Err(err).Msg("Unable to parse environment variables.")
	}

	logger.Info().Bool("UsePostgres", env.UsePostgres).
		Str("PostgresHostname", env.PostgresHostname).
		Uint32("PostgresPort", env.PostgresPort).
		Str("PostgresDBName", env.PostgresDBName).
		Str("PostgresUser", env.PostgresUser).
		Msg("Backend starts. 🚀")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := StartWebRouter(ctx, env); err != nil {
		// * Sleep 5 seconds
		time.Sleep(time.Second * 5)

		logger.Fatal().Err(err).Msg("StartWebRouter")
	}
}

func StartWebRouter(ctx context.Context, env base.Env) error {

	router := gin.Default()

	if env.UsePostgres {
		return runInPostgres(ctx, router, env)
	} else {
		return runInMemory(router)
	}
}

func runInMemory(router *gin.Engine) error {
	memDB := memory.NewMemoryDB()

	loanRepo := memoryDB.NewLoanRepository(memDB)

	loanUC := usecase.NewLoanUsecase(loanRepo)

	http.NewLoanHandler(router, loanUC)

	router.Run(":8080")

	return nil
}

func runInPostgres(ctx context.Context, router *gin.Engine, env base.Env) error {
	logger := log.NewLogger()

	database, err := postgres.NewDatabase(
		ctx,
		postgres.DatabaseParams{
			Hostname: env.PostgresHostname,
			Port:     env.PostgresPort,
			Name:     env.PostgresDBName,
			Timeout:  env.PostgresTimeout,
		},
		postgres.DatabaseCredentials{
			Username: env.PostgresUser,
			Password: env.PostgresPassword,
		},
	)
	if err != nil {
		logger.Error().Err(err).Str("database", env.PostgresHostname).
			Uint32("port", env.PostgresPort).
			Msg("unable to connect to database")
		return err
	}

	defer database.Close()

	loanRepo := postgresDB.NewLoanRepository(database)

	loanUC := usecase.NewLoanUsecase(loanRepo)

	http.NewLoanHandler(router, loanUC)

	router.Run(":8080")

	return nil
}
