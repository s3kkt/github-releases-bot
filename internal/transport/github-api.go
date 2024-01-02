package transport

import (
	"encoding/json"
	"errors"
	zlog "github.com/rs/zerolog/log"
	"github.com/s3kkt/github-releases-bot/internal"
	"io"
	"net/http"
)

func GetReleases(repoUrl string, token string) (internal.Release, error) {
	client := &http.Client{}
	release := internal.Release{}
	apiError := internal.APIError{}

	req, err := http.NewRequest("GET", repoUrl, nil)
	if err != nil {
		zlog.Error().Msgf("API request failed: %s", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := client.Do(req)
	if err != nil {
		zlog.Error().Msgf("API request failed: %s", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		zlog.Error().Msgf("Cannot read API responce: %s", err)
	}

	err = json.Unmarshal(body, &release)
	if err != nil {
		zlog.Error().Msgf("Cannot unmashal API responce JSON: %s", err)
	}

	if release.HtmlUrl == "" {
		err = json.Unmarshal(body, &apiError)
		if err != nil {
			zlog.Error().Msgf("Cannot unmashal API responce JSON: %s", err)
		}
		err := errors.New("Error recieving latest release for " + repoUrl + " Message: " + apiError.Message)
		return release, err
	} else {
		return release, nil
	}
}
