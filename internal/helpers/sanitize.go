package helpers

import (
	"log"
	"strings"
)

func ReposListOutput(reposList []string) string {
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
