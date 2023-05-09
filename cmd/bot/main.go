package main

import (
	"github.com/s3kkt/github-releases-bot/internal/config"
	"github.com/s3kkt/github-releases-bot/internal/database"
	"github.com/s3kkt/github-releases-bot/internal/telegram"
	"os"
)

var (
	configPath = config.ParseFlags()
	db         = config.DatabaseConnectionString(configPath)
	conf       = config.GetConfig(configPath)
)

func main() {
	err := os.Setenv("DB_CONNECTION_STRING", db)
	if err != nil {
		return
	}

	database.CheckDatabaseConnection()
	database.Cleanup(conf)

	for _, r := range conf.RepoUrl {
		database.AddRepo(r, true, 0)
	}

	//go database.Updater(conf)
	go telegram.Notifier(conf)

	telegram.Bot(conf)
}
