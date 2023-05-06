package main

import (
	"fmt"
	"github.com/s3kkt/github-releases-bot/internal/config"
	"github.com/s3kkt/github-releases-bot/internal/repository"
	"github.com/s3kkt/github-releases-bot/internal/transport"
	"time"
)

func main() {
	configPath := config.ParseFlags()
	conf := config.GetConfig(configPath)
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		conf.Database.Username, conf.Database.Password, conf.Database.Host, conf.Database.Port, conf.Database.DBName)

	repository.CheckDatabaseConnection(connectionString)

	for _, v := range conf.RepoUrl {
		release := transport.GetReleases(config.GetApiURL(v), conf.GitHubToken)
		check_time := time.Now().Format(time.RFC3339)

		repository.InsertRepoFromConfig(connectionString, v)
		repository.InsertReleaseData(connectionString, check_time, v, release)
	}
}
