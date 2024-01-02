package main

import (
	"database/sql"
	"github.com/heptiolabs/healthcheck"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/s3kkt/github-releases-bot/internal"
	"github.com/s3kkt/github-releases-bot/internal/config"
	"github.com/s3kkt/github-releases-bot/internal/database"
	"github.com/s3kkt/github-releases-bot/internal/telegram"
	"net/http"
	"os"
	"time"
)

var (
	configPath,
	dbMigrate,
	runInCloud = config.ParseFlags()
)

func main() {

	var cfg internal.Config

	if configPath != "" {
		config.ReadConfigFile(configPath, &cfg)
	}
	config.ReadConfigEnv(&cfg)

	logLevel := config.ParseLogLevel(cfg.LogLevel)
	zerolog.SetGlobalLevel(logLevel)

	zlog.Info().Msg("Starting bot...")

	dbString := config.DatabaseConnectionString(cfg, runInCloud)

	health := healthcheck.NewHandler()

	health.AddLivenessCheck("goroutine-threshold", healthcheck.GoroutineCountCheck(100))
	health.AddReadinessCheck(
		"upstream-dep-dns",
		healthcheck.DNSResolveCheck("api.telegram.org", 50*time.Millisecond))

	go func() {
		err := http.ListenAndServe("0.0.0.0:8080", health)
		if err != nil {
			zlog.Fatal().Msgf("Cannot open heathcheck port: ", err)
		}
	}()

	err := os.Setenv("DB_CONNECTION_STRING", dbString)
	if err != nil {
		return
	}

	db, err := sql.Open("postgres", dbString)
	if err != nil {
		zlog.Fatal().Msgf("Cannot connect database: %s", err)
	}

	defer db.Close()

	dbx, err := sqlx.Open("postgres", dbString)
	if err != nil {
		zlog.Fatal().Msgf("Cannot connect database: %s", err)
	}

	defer dbx.Close()

	githubRepo := database.NewGithubRepo(db, dbx)

	githubRepo.CheckDatabaseConnection()

	if dbMigrate == true {
		err = githubRepo.Migrate()
		if err != nil {
			zlog.Fatal().Msgf("Migration failed! Reason: %s", err)
		} else {
			zlog.Info().Msgf("Migrations successfully applied. Application stopped.")
			os.Exit(0)
		}
	}

	tg := telegram.NewTG(githubRepo)

	_, chatsList := githubRepo.GetChatsList()
	for i := range chatsList {
		go tg.Notifier(cfg, chatsList[i])
	}

	tg.Bot(cfg)
}
