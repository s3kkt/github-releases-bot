package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/s3kkt/github-releases-bot/internal"
	"log"
)

type GithubRepos struct {
	db  *sql.DB
	dbx *sqlx.DB
}

func NewGithubRepo(db *sql.DB, dbx *sqlx.DB) *GithubRepos {
	return &GithubRepos{
		db:  db,
		dbx: dbx,
	}
}

func (r *GithubRepos) Migrate() error {
	driver, err := postgres.WithInstance(r.db, &postgres.Config{})
	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/database/migrations",
		"postgres", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil {
		return err
	}

	return nil
}

func (r *GithubRepos) CheckDatabaseConnection() bool {

	if err := r.db.Ping(); err != nil {
		log.Fatal("Database connection failed! Reason: ", err)
	}

	log.Println("Successfully connected to database!")
	return true
}

func (r *GithubRepos) UpdateChat(chatId int64, userName, firstName, lastName, chatType string, isBot bool, date int64) error {
	sqlStatement := `
	INSERT INTO chats (chat_id, username, first_name, last_name, type, is_bot, last_activity)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    ON CONFLICT (chat_id, username) DO UPDATE SET last_activity = $7;`

	_, err := r.db.Exec(sqlStatement, chatId, userName, firstName, lastName, chatType, isBot, date)
	log.Printf("Updating chat: %d.", chatId)
	if err != nil {
		log.Printf("Database chat update failed. ChatID: %d Reason: %s", chatId, err)
	}

	return err
}

func (r *GithubRepos) AddRepo(repo string, chatId int64) {
	sqlStatement := `
	INSERT INTO repos (name, chat_id)
        VALUES ($1, $2)
    ON CONFLICT (name, chat_id) DO UPDATE SET deleted = false, chat_id = $2;`

	_, err := r.db.Exec(sqlStatement, repo, chatId)
	log.Printf("Adding repo: %s.", repo)
	if err != nil {
		log.Printf("Failed add repo %s. ChatID: %d Reason: %s", repo, chatId, err)
		//return
	}

	return
}

func (r *GithubRepos) DeleteRepo(repo string, chatId int64) error {
	sqlStatement := `UPDATE repos SET deleted = true WHERE name = $1 AND chat_id = $2;`

	_, err := r.db.Exec(sqlStatement, repo, chatId)
	log.Print("Disabling unused repo: ", repo)
	if err != nil {
		log.Fatal("Repo deleting failed. Reason: ", err)
	}

	return nil
}

func (r *GithubRepos) GetChatsList() (error, []int64) {
	sqlStatement := `
    SELECT DISTINCT chat_id
    FROM chats WHERE (SELECT extract(epoch from now())) - last_activity < 60*60*24*30*6;`

	var list []int64

	rows, err := r.db.Query(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	for rows.Next() {
		var r int64
		if err := rows.Scan(&r); err != nil {
			return err, list
		}
		list = append(list, r)
	}
	if err = rows.Err(); err != nil {
		return err, list
	}

	return nil, list
}

func (r *GithubRepos) GetChatReposList(chatId int64) (error, []string) {
	sqlStatement := `SELECT name FROM repos WHERE deleted != true and chat_id = $1;`
	var list []string

	rows, err := r.db.Query(sqlStatement, chatId)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err != nil {
			return err, list
		}
		list = append(list, r)
	}
	if err = rows.Err(); err != nil {
		return err, list
	}

	return nil, list
}

func (r *GithubRepos) GetChatLatestList(chatId int64) ([]internal.LatestRelease, error) {
	var latestReleaseList []internal.LatestRelease

	sqlStatement := `SELECT 
        releases.repo_name,
        releases.tag_name,
        releases.published_at
    FROM releases LEFT JOIN repos ON repos.name = releases.repo_name
    WHERE repos.chat_id = $1 AND latest = true`

	//err = db.Select(&latestReleaseList, "select releases.repo_name, releases.tag_name, releases.published_at from releases left join repos on repos.name = releases.repo_name where repos.chat_id = $1 and latest = true", chatId)
	err := r.dbx.Select(&latestReleaseList, sqlStatement, chatId)
	if err != nil {
		return nil, err
	}
	//err = db.Select(&pp, "select repo_name, tag_name, published_at from releases")
	//log.Printf("DEBUG: latest release for chat %v: %v\n", chatId, latestRelease)

	return latestReleaseList, nil
}

func (r *GithubRepos) InsertReleaseData(checkTime, repo string, release internal.Release, latest bool) {
	sqlStatement := `
	INSERT INTO releases (
	    repo_name,
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
	    last_check,
	    latest
	    )
    VALUES ($1, $2, $3, $4, $5, $6, $7,$8, $9, $10, $11, $12, $13, $14, $15);`

	tx := r.dbx.MustBegin()
	tx.MustExec(sqlStatement,
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
		checkTime,
		latest,
	)
	tx.MustExec("UPDATE releases SET latest = false WHERE repo_name = $1 AND tag_name <> $2", repo, release.TagName)
	err := tx.Commit()
	if err != nil {
		log.Printf("Inserting release data failed. Reason:", err)
		return
	}

	return
}

func (r *GithubRepos) CheckIfNew(repo, tag string) (bool, error) {
	sqlStatement := `
    SELECT
        releases.repo_name,
        releases.tag_name
    FROM releases
    LEFT JOIN repos ON repos.name = releases.repo_name
    WHERE 
        repo_name = $1 AND
        tag_name = $2 AND
        repos.deleted = false;`

	rows, err := r.dbx.Query(sqlStatement, repo, tag)
	if err != nil {
		log.Fatal("Check failed. Reason: ", err)
	}

	defer rows.Close()

	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err != nil {
			newError := fmt.Sprintf("release %s for %s already added to database", tag, repo)
			e := errors.New(newError)
			return false, e
		}
	}
	if err = rows.Err(); err != nil {
		return false, err
	}

	return true, nil
}
