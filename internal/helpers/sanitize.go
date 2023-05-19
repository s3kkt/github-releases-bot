package helpers

import (
	"log"
	"regexp"
	"strings"
)

func ReposListOutput(reposList []string) string {
	if len(reposList) == 0 {
		return "There is no repos at this moment."
	}
	return strings.Join(reposList, "\n")
}

func ValidateRepoUrl(repoUrl string) bool {
	if strings.HasPrefix(repoUrl, "https://github.com") == true {
		log.Printf("Repo format validation successful for %s", repoUrl)
		return true
	} else {
		log.Printf("Repo format validation failed for %s. Must be a 'https://github.com/author/repo'", repoUrl)
		return false
	}
}

func SanitizeRepoName(repo string) string {
	re, err := regexp.Compile(`https:\/\/github.com\/`)
	if err != nil {
		log.Fatal(err)
	}
	repo = re.ReplaceAllString(repo, "")
	return repo
}

func SanitizeReleaseNotes(releaseNotes string) string {
	unsupportedRegex := [...]string{
		`<`,
		`>`,
	}
	for r := range unsupportedRegex {
		re, err := regexp.Compile(unsupportedRegex[r])
		if err != nil {
			log.Fatal(err)
		}
		releaseNotes = re.ReplaceAllString(releaseNotes, "")
	}
	if len(releaseNotes) > 300 {
		return releaseNotes[:300] + "\n...\n"
	}

	return releaseNotes
}
