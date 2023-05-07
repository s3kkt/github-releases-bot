package internal

import "time"

type Config struct {
	GitHubToken    string   `yaml:"github_token"`
	TelegramToken  string   `yaml:"telegram_token"`
	Debug          bool     `yaml:"debug"`
	UpdateInterval string   `yaml:"update_interval"`
	RepoUrl        []string `yaml:"repos"`
	Database       struct {
		Username string `yaml:"user"`
		Password string `yaml:"pass"`
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		DBName   string `yaml:"dbname"`
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
