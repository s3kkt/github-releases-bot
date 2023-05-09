package database

import (
	"database/sql"
	"errors"
	"fmt"
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

func AddRepo(repo string, fromConfig bool, chatId int64) {
	connectionString := os.Getenv("DB_CONNECTION_STRING")

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal("Add repo failed. Reason: ", err)
	}

	if fromConfig == false {
		sqlStatement := `
	        INSERT INTO repos (name, from_config, chat_id)
                VALUES ($1, true, $3)
            ON CONFLICT (name) DO UPDATE SET deleted = false, from_config = $2, chat_id = $3;`

		_, err = db.Exec(sqlStatement, repo, fromConfig, chatId)
		log.Printf("Adding repo: %s. From config: %+v", repo, fromConfig)
		if err != nil {
			log.Printf("Failed add repo %s. From config: %v, ChatID: %d Reason :%s", repo, fromConfig, chatId, err)
			return
		}
	} else {
		sqlStatement := `
	        INSERT INTO repos (name, from_config)
                VALUES ($1, $2)
            ON CONFLICT (name) DO UPDATE SET deleted = false, from_config = $2;`

		_, err = db.Exec(sqlStatement, repo, fromConfig)
		log.Printf("Adding repo: %s. From config: %+v", repo, fromConfig)
		if err != nil {
			log.Printf("Failed add repo %s. Reason :%s", repo, err)
		}
	}

	defer db.Close()

	return
}

func Cleanup(conf internal.Config) {
	connectionString := os.Getenv("DB_CONNECTION_STRING")
	var l []interface{}
	var k []string
	for idx := range conf.RepoUrl {
		k = append(k, fmt.Sprintf("$%d", idx+1))
	}
	for _, r := range conf.RepoUrl {
		l = append(l, r)
	}

	//log.Println(strings.Join(l, ","))
	sqlStatement := fmt.Sprintf("UPDATE repos SET deleted = true WHERE name NOT IN (%s) AND from_config = true;", strings.Join(k, ","))
	fmt.Println(sqlStatement)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(sqlStatement, l...)
	log.Print("Disabling unused repos...")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	return
}

func DeleteRepo(repo string) error {
	connectionString := os.Getenv("DB_CONNECTION_STRING")
	sqlStatement := `UPDATE repos SET deleted = true WHERE name = $1;`

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	if CheckFromConfig(repo) == false {
		_, err = db.Exec(sqlStatement, repo)
		log.Print("Disabling unused repo: ", repo)
		if err != nil {
			log.Fatal("Repo deleting failed. Reason: ", err)
		}

	} else {
		log.Printf("Repo deleting failed for %s. Reason: repo from config", repo)
		return nil
	}

	defer db.Close()

	return nil
}

func GetReposList() (error, []string) {
	connectionString := os.Getenv("DB_CONNECTION_STRING")
	sqlStatement := `SELECT name FROM repos WHERE deleted != true;`
	var list []string

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

func InsertReleaseData(checkTime, repo string, release internal.Release) map[string]string {
	connectionString := os.Getenv("DB_CONNECTION_STRING")
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
    ON CONFLICT (repo_name, tag_name) DO UPDATE SET last_check = $14 
    RETURNING repo_name, author, tag_name, release_url, target_branch, release_text, published_at, is_draft, is_prerelease;`

	var repoName, author, tagName, releaseUrl, targetBranch, releaseNotes string
	var isdraft, isprerelease bool
	var publishedAt time.Time
	newRelease := make(map[string]string)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal("Inserting release data failed. Reason:", err)
	}
	err = db.QueryRow(sqlStatement,
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
		checkTime).Scan(&repoName, &author, &tagName, &releaseUrl, &targetBranch, &releaseNotes, &publishedAt, &isdraft, &isprerelease)
	if err != nil {
		log.Fatal("Inserting release data failed. Reason: ", err)
	} else {
		// Пишет в лог N раз, потому что параметры возвращаются при каждом insert из цикла Updater()
		// Написать функцию чекер, которая будет выполнять проверку перед инсертом и вставлять данные только если релиз действительно новый
		// Можно будет убрать ON CONFLICT и UNIQUE с полей в таблице releases

		// Добавлен чекер, нужно проверить, можно ли убрать ON CONFLICT и UNIQUE с полей в таблице releases
		_, err := CheckIfNew(repoName, tagName, isdraft, isprerelease)
		//if err != nil {
		//	return nil
		//}
		if err != nil {
			log.Println("Inserting release data failed. Reason: ", err)
			return nil
		} else {
			fmt.Println("New release!!!")
			fmt.Println("Repo name:", repoName)
			fmt.Println("Author:", author)
			fmt.Println("Tag:", tagName)
			fmt.Println("Release URL:", releaseUrl)
			fmt.Println("Branch", targetBranch)
			fmt.Println("Notes", targetBranch)
			fmt.Println("Published at:", publishedAt.String())

			//var p = internal.NewRelease {
			//    RepoName:   repoName,
			//	Author:     author,
			//	Tag:        tagName,
			//	ReleaseURL: releaseUrl,
			//	Branch:     targetBranch,
			//	Date:       publishedAt.String(),
			//}

			newRelease["RepoName"] = repoName
			newRelease["Author"] = author
			newRelease["Tag"] = tagName
			newRelease["ReleaseURL"] = releaseUrl
			newRelease["Branch"] = targetBranch
			newRelease["Text"] = releaseNotes
			newRelease["Date"] = publishedAt.String()
		}
	}

	defer db.Close()

	return newRelease
}

func CheckIfNew(repo, tag string, isdraft, isprrelease bool) (bool, error) {
	connectionString := os.Getenv("DB_CONNECTION_STRING")
	sqlStatement := `
    SELECT repo_name, tag_name, is_draft, is_prerelease 
    FROM releases
    WHERE 
        repo_name = $1 AND
        tag_name = $2 AND
        is_draft = $3 AND
        is_prerelease = $4;`

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query(sqlStatement, repo, tag, isdraft, isprrelease)
	if err != nil {
		log.Fatal("Check failed. Reason: ", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r string
		if err := rows.Scan(&r); err != nil {
			newError := fmt.Sprintf("release %s already added to database", repo)
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

func CheckFromConfig(repo string) bool {
	connectionString := os.Getenv("DB_CONNECTION_STRING")
	sqlStatement := `SELECT from_config FROM repos WHERE name = $1;`
	var fromConfig bool

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query(sqlStatement, repo)
	if err != nil {
		log.Fatal("Check if repo added from config failed. Reason: ", err)
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&fromConfig); err != nil {
			fmt.Println("Check if repo added from config failed. Reason: ", err)
		}
	}
	if err = rows.Err(); err != nil {
		fmt.Println("Check if repo added from config failed. Reason: ", err)
	}

	defer db.Close()

	return fromConfig
}

func Updater(conf internal.Config) {
	duration, _ := time.ParseDuration(conf.UpdateInterval)
	for range time.Tick(duration) {
		log.Print("Check for updates...")
		for _, repo := range conf.RepoUrl {
			release, err := transport.GetReleases(config.GetApiURL(repo), conf.GitHubToken)
			if err != nil {
				log.Println(err)
				return
			}
			checkTime := time.Now().Format(time.RFC3339)
			InsertReleaseData(checkTime, repo, release)
		}
	}
	return
}
