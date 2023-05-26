package internal

import (
	_ "github.com/kelseyhightower/envconfig"
	"time"
)

type Config struct {
	GitHubToken    string `yaml:"github_token" envconfig:"BOT_GITHUB_TOKEN"`
	TelegramToken  string `yaml:"telegram_token" envconfig:"BOT_TELEGRAM_TOKEN"`
	Debug          bool   `yaml:"debug" envconfig:"BOT_DEBUG"`
	UpdateInterval string `yaml:"update_interval" envconfig:"BOT_UPDATE_INTERVAL"`
	Database       struct {
		Username string `yaml:"user" envconfig:"BOT_DB_USER"`
		Password string `yaml:"pass" envconfig:"BOT_DB_PASS"`
		Host     string `yaml:"host" envconfig:"BOT_DB_HOST"`
		Port     int    `yaml:"port" envconfig:"BOT_DB_PORT"`
		DBName   string `yaml:"dbname" envconfig:"BOT_DB_NAME"`
	} `yaml:"database"`
}

type Release struct {
	HtmlUrl         string    `json:"html_url"`
	TagName         string    `json:"tag_name"`
	TargetCommitish string    `json:"target_commitish"`
	Name            string    `json:"name"`
	Draft           bool      `json:"draft"`
	Prerelease      bool      `json:"prerelease"`
	CreatedAt       time.Time `json:"created_at"`
	PublishedAt     time.Time `json:"published_at"`
	TarballUrl      string    `json:"tarball_url"`
	ZipballUrl      string    `json:"zipball_url"`
	Body            string    `json:"body"`
	Author          struct {
		Login string `json:"login"`
	} `json:"author"`
}

type LatestRelease struct {
	RepoName    string    `db:"repo_name"`
	TagName     string    `db:"tag_name"`
	PublishedAt time.Time `db:"published_at"`
}

type APIError struct {
	Message          string `json:"message"`
	DocumentationUrl string `json:"documentation_url"`
}

type LetestList struct {
	Tag  string `ddb:"tag_name"`
	Repo int    `ddb:"repo_name"`
}
