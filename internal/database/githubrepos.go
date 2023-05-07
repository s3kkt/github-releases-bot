package database

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/s3kkt/github-releases-bot/internal"
	"github.com/s3kkt/github-releases-bot/internal/config"
	"github.com/s3kkt/github-releases-bot/internal/transport"
	"log"
	"os"
	"strings"
	"time"
)

func CheckDatabaseConnection() bool {
	connectionString := os.Getenv("DB_CONNECTION_STRING")
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Database connection failed!", err)
	}

	log.Println("Successfully connected to database!")
	return true
}

func AddRepoFromConfig(conf internal.Config) {
	connectionString := os.Getenv("DB_CONNECTION_STRING")
	sqlStatement := `
	INSERT INTO repos (id, name, from_config)
        VALUES (DEFAULT, $1, true)
    ON CONFLICT (name) DO UPDATE SET 
        id = (SELECT id FROM repos WHERE name = $1),
        name = $1,
        from_config = true;`

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	for _, r := range conf.RepoUrl {
		_, err = db.Exec(sqlStatement, r)
		log.Printf("Add repo from config: %s", r)
		if err != nil {
			log.Fatal(err)
		}
	}
	defer db.Close()

	return
}

func Cleanup(conf internal.Config) {
	connectionString := os.Getenv("DB_CONNECTION_STRING")
	var l []string
	for _, r := range conf.RepoUrl {
		l = append(l, r)
	}
	list := "'" + strings.Join(l, "','") + "'"

	sqlStatement := `DELETE FROM repos WHERE name NOT IN ($1) AND from_config != false`

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(sqlStatement, list)
	log.Print("Cleanup not listed repos...")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	return
}

func GetReposList() ([]string, error) {
	connectionString := os.Getenv("DB_CONNECTION_STRING")
	sqlStatement := `SELECT name FROM repos;`
	var list []string

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
		return list, err
	}

	rows, err := db.Query(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err != nil {
			return list, err
		}
		list = append(list, r)
	}
	if err = rows.Err(); err != nil {
		return list, err
	}

	defer db.Close()

	return list, nil
}

func InsertReleaseData(checkTime, repo string, release internal.Release) {
	connectionString := os.Getenv("DB_CONNECTION_STRING")
	sqlStatement := `
	INSERT INTO releases (
	      repo_id,
	      release_url,
	      author,
	      tag_name,
	      release_name,
	      target_branch,
	      is_draft,
	      is_prerelease,
	      tarball_url,
	      zipball_url,
	      release_text,
	      created_at,
	      published_at,
	      last_check
	      )
    VALUES ((SELECT id FROM repos WHERE name = $1), $2, $3, $4, $5, $6, $7,$8, $9, $10, $11, $12, $13, $14)
    ON CONFLICT (repo_id, tag_name) DO NOTHING;`

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(sqlStatement,
		repo,
		release.HtmlUrl,
		release.Author.Login,
		release.TagName,
		release.Name,
		release.TargetCommitish,
		release.Draft,
		release.Prerelease,
		release.TarballUrl,
		release.ZipballUrl,
		release.Body,
		release.CreatedAt,
		release.PublishedAt,
		checkTime)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	return
}

func Updater(conf internal.Config) {
	duration, _ := time.ParseDuration(conf.UpdateInterval)
	for range time.Tick(duration) {
		log.Print("Check for updates...")
		for _, v := range conf.RepoUrl {
			release := transport.GetReleases(config.GetApiURL(v), conf.GitHubToken)
			checkTime := time.Now().Format(time.RFC3339)
			InsertReleaseData(checkTime, v, release)
		}
	}
	return
}
