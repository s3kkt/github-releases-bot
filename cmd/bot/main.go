package main

import (
	"github.com/s3kkt/github-releases-bot/internal/config"
	"github.com/s3kkt/github-releases-bot/internal/database"
	"github.com/s3kkt/github-releases-bot/internal/telegram"
	"os"
)

var (
	configPath = config.ParseFlags()
	dbString   = config.DatabaseConnectionString(configPath)
	conf       = config.GetConfig(configPath)
)

func main() {

	//m, err := migrate.New()
	//m.steps(2)

	err := os.Setenv("DB_CONNECTION_STRING", dbString)
	if err != nil {
		return
	}

	database.CheckDatabaseConnection()

	go telegram.Notifier(conf)
	telegram.Bot(conf)
}
