package config

import (
	"flag"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/s3kkt/github-releases-bot/internal"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
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
		zlog.Fatal().Msgf("Cannot open config file: %s", err)
	}

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		zlog.Fatal().Msgf("Cannot decode config: %s", err)
	}
}

func ReadConfigEnv(cfg *internal.Config) {
	err := envconfig.Process("", cfg)
	if err != nil {
		zlog.Fatal().Msgf("Cannot read environment variables: %s", err)
	}
}

func ParseLogLevel(level string) zerolog.Level {
	var l zerolog.Level
	switch {
	case strings.EqualFold(level, "debug"):
		l = 0
	case strings.EqualFold(level, "info"):
		l = 1
	case strings.EqualFold(level, "warn"):
		l = 2
	case strings.EqualFold(level, "error"):
		l = 3
	case strings.EqualFold(level, "fatal"):
		l = 4
	case strings.EqualFold(level, "trace"):
		l = -1
	}
	return l
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
