package main

import (
	"github.com/heptiolabs/healthcheck"
	"github.com/s3kkt/github-releases-bot/internal"
	"github.com/s3kkt/github-releases-bot/internal/config"
	"github.com/s3kkt/github-releases-bot/internal/database"
	"github.com/s3kkt/github-releases-bot/internal/telegram"
	"log"
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

	dbString := config.DatabaseConnectionString(cfg, runInCloud)

	health := healthcheck.NewHandler()

	health.AddLivenessCheck("goroutine-threshold", healthcheck.GoroutineCountCheck(100))
	health.AddReadinessCheck(
		"upstream-dep-dns",
		healthcheck.DNSResolveCheck("api.telegram.org", 50*time.Millisecond))

	go func() {
		err := http.ListenAndServe("0.0.0.0:8080", health)
		if err != nil {
			log.Fatal("Cannot open heathcheck port: ", err)
		}
	}()

	err := os.Setenv("DB_CONNECTION_STRING", dbString)
	if err != nil {
		return
	}

	database.CheckDatabaseConnection()

	if dbMigrate == true {
		err = database.Migrate()
		if err != nil {
			log.Fatalf("Migration failed! Reason: %s", err)
		} else {
			log.Println("Migrations successfully applied")
		}
	}

	_, chatsList := database.GetChatsList()
	for i := range chatsList {
		go telegram.Notifier(cfg, chatsList[i])
	}

	telegram.Bot(cfg)
}
