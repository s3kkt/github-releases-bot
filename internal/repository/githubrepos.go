package repository

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/s3kkt/github-releases-bot/internal"
	"log"
)

func CheckDatabaseConnection(connectionString string) bool {
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

func InsertRepoFromConfig(connectionString, name string) {
	sqlStatement := `
	INSERT INTO repos (id, name)
       VALUES (DEFAULT, $1)
       ON CONFLICT (name) DO NOTHING;`

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(sqlStatement, name)
	log.Printf("Try to add repo: %s", name)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}

func InsertReleaseData(connectionString, check_time, repo string, release internal.Release) {
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
    VALUES (
            (SELECT id FROM repos WHERE name = $1),
            $2, $3, $4, $5, $6, $7,$8, $9, $10, $11, $12, $13, $14
            )
    ON CONFLICT (repo_id, tag_name) DO NOTHING;`

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(sqlStatement, repo,
		release.HtmlUrl, release.Author.Login, release.TagName, release.Name,
		release.TargetCommitish, release.Draft, release.Prerelease,
		release.TarballUrl, release.ZipballUrl, release.Body,
		release.CreatedAt, release.PublishedAt, check_time)
	//log.Printf("Try to add repo: %s", name)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
}
