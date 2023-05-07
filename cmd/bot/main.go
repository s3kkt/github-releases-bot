package main

import (
	"github.com/s3kkt/github-releases-bot/internal/config"
	"github.com/s3kkt/github-releases-bot/internal/database"
	"github.com/s3kkt/github-releases-bot/internal/telegram"
	"os"
)

func main() {
	configPath := config.ParseFlags()
	db := config.DatabaseConnectionString(configPath)
	conf := config.GetConfig(configPath)

	err := os.Setenv("DB_CONNECTION_STRING", db)
	if err != nil {
		return
	}

	database.CheckDatabaseConnection()
	database.Cleanup(conf)
	database.AddRepoFromConfig(conf)
	go database.Updater(conf)
	telegram.Bot(conf)
}
