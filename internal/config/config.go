package config

import (
	"flag"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/s3kkt/github-releases-bot/internal"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

func ParseFlags() (string, bool, bool) {
	configPath := flag.String("config", "", "path to YAML config file")
	dbMigrate := flag.Bool("migrations", false, "set true to run migrations only")
	runInCloud := flag.Bool("cloud", false, "set true to run in GCP")
	flag.Parse()
	return *configPath, *dbMigrate, *runInCloud
}

func ReadConfigFile(configPath string, cfg *internal.Config) {
	f, err := os.Open(configPath)
	if err != nil {
		log.Fatal(err)
	}

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		log.Fatal(err)
	}
}

func ReadConfigEnv(cfg *internal.Config) {
	err := envconfig.Process("", cfg)
	if err != nil {
		log.Fatal(err)
	}
}

func DatabaseConnectionString(conf internal.Config, runInCloud bool) string {
	var conn string
	if runInCloud == true {
		conn = fmt.Sprintf(
			"host=/cloudsql/%s user=%s password=%s dbname=%s sslmode=disable",
			conf.Database.Host,
			conf.Database.Username,
			conf.Database.Password,
			conf.Database.DBName)
	} else {
		conn = fmt.Sprintf(
			"postgres://%v:%v@%v:%v/%v?sslmode=disable",
			conf.Database.Username,
			conf.Database.Password,
			conf.Database.Host,
			conf.Database.Port,
			conf.Database.DBName,
		)
	}
	return conn
}
