package config

import (
	"flag"
	"fmt"
	"github.com/s3kkt/github-releases-bot/internal"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"regexp"
)

func ParseFlags() string {
	configPath := flag.String("config", "/", "path to YAML config file")
	flag.Parse()
	return *configPath
}

func GetConfig(configPath string) internal.Config {
	config, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal(err)
	}
	var c internal.Config

	if err := yaml.Unmarshal(config, &c); err != nil {
		log.Fatal(err)
	}
	return c
}

func GetApiURL(url string) string {
	re := regexp.MustCompile(`github.com/`)
	return re.ReplaceAllString(url, `api.github.com/repos/`) + "/releases/latest"
}

func DatabaseConnectionString(configPath string) string {
	conf := GetConfig(configPath)
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		conf.Database.Username, conf.Database.Password, conf.Database.Host, conf.Database.Port, conf.Database.DBName)
}
