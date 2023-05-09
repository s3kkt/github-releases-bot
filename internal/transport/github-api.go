package transport

import (
	"encoding/json"
	"fmt"
	"github.com/s3kkt/github-releases-bot/internal"
	"io"
	"log"
	"net/http"
)

func GetReleases(repoUrl string, token string) internal.Release {
	client := &http.Client{}
	req, err := http.NewRequest("GET", repoUrl, nil)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	var release internal.Release
	err = json.Unmarshal(body, &release)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	return release
}
