package config

import (
	"flag"
	"github.com/s3kkt/github-releases-bot/internal"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"regexp"
)

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

func ParseFlags() string {
	configPath := flag.String("config", "/", "path to YAML config file")
	flag.Parse()
	return *configPath
	//return "./configs/config.yml"
}

func GetApiURL(url string) string {
	re := regexp.MustCompile(`github.com/`)
	return re.ReplaceAllString(url, `api.github.com/repos/`) + "/releases/latest"
}
