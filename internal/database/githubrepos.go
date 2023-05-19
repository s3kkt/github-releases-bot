package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/s3kkt/github-releases-bot/internal"
	"log"
	"os"
)

func Migrate() error {
	db, err := sql.Open("postgres", os.Getenv("DB_CONNECTION_STRING"))
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
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

func CheckDatabaseConnection() bool {
	connectionString := os.Getenv("DB_CONNECTION_STRING")
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Database connection failed! Reason: ", err)
	}

	log.Println("Successfully connected to database!")
	return true
}

func UpdateChat(chatId int64, userName, firstName, lastName, chatType string, isBot bool, date int64) error {
	connectionString := os.Getenv("DB_CONNECTION_STRING")

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal("Database chat update failed. Reason: ", err)
	}

	sqlStatement := `
	INSERT INTO chats (chat_id, username, first_name, last_name, type, is_bot, last_activity)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    ON CONFLICT (chat_id, username) DO UPDATE SET last_activity = $7;`

	_, err = db.Exec(sqlStatement, chatId, userName, firstName, lastName, chatType, isBot, date)
	log.Printf("Updating chat: %d.", chatId)
	if err != nil {
		log.Printf("Database chat update failed. ChatID: %d Reason: %s", chatId, err)
	}

	defer db.Close()

	return err
}

func AddRepo(repo string, chatId int64) {
	connectionString := os.Getenv("DB_CONNECTION_STRING")

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal("Add repo failed. Reason: ", err)
	}

	sqlStatement := `
	INSERT INTO repos (name, chat_id)
        VALUES ($1, $2)
    ON CONFLICT (name, chat_id) DO UPDATE SET deleted = false, chat_id = $2;`

	_, err = db.Exec(sqlStatement, repo, chatId)
	log.Printf("Adding repo: %s.", repo)
	if err != nil {
		log.Printf("Failed add repo %s. ChatID: %d Reason: %s", repo, chatId, err)
		//return
	}

	defer db.Close()

	return
}

func DeleteRepo(repo string, chatId int64) error {
	connectionString := os.Getenv("DB_CONNECTION_STRING")
	sqlStatement := `UPDATE repos SET deleted = true WHERE name = $1 AND chat_id = $2;`

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(sqlStatement, repo, chatId)
	log.Print("Disabling unused repo: ", repo)
	if err != nil {
		log.Fatal("Repo deleting failed. Reason: ", err)
	}

	defer db.Close()

	return nil
}

func GetChatsList() (error, []int64) {
	connectionString := os.Getenv("DB_CONNECTION_STRING")
	sqlStatement := `
    SELECT DISTINCT chat_id
    FROM chats WHERE (SELECT extract(epoch from now())) - last_activity < 60*60*24*30*6;`

	var list []int64

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
		return err, list
	}

	rows, err := db.Query(sqlStatement)
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

	defer db.Close()

	return nil, list
}

func GetChatReposList(chatId int64) (error, []string) {
	connectionString := os.Getenv("DB_CONNECTION_STRING")
	sqlStatement := `SELECT name FROM repos WHERE deleted != true and chat_id = $1;`
	var list []string

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
		return err, list
	}

	rows, err := db.Query(sqlStatement, chatId)
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

	defer db.Close()

	return nil, list
}

func InsertReleaseData(checkTime, repo string, release internal.Release) {
	connectionString := os.Getenv("DB_CONNECTION_STRING")
	// After checking in Notifier, call InsertReleaseData occurs only on new release and there is no conflicts now.
	// ToDo: Remove last_check field and ON CONFLICT condition
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
	    last_check
	    )
    VALUES ($1, $2, $3, $4, $5, $6, $7,$8, $9, $10, $11, $12, $13, $14)
    ON CONFLICT (repo_name, tag_name) DO UPDATE SET last_check = $14;`

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal("Inserting release data failed. Reason:", err)
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
		log.Fatal("Query for inserting latest release data failed. Reason: ", err)
	}

	defer db.Close()

	return
}

func CheckIfNew(repo, tag string) (bool, error) {
	connectionString := os.Getenv("DB_CONNECTION_STRING")
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

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query(sqlStatement, repo, tag)
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

	defer db.Close()

	return true, nil
}
