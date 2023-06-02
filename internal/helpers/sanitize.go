package helpers

import (
	"fmt"
	"github.com/s3kkt/github-releases-bot/internal"
	"log"
	"regexp"
	"strings"
)

func GetApiURL(url string) string {
	re := regexp.MustCompile(`github.com/`)
	return re.ReplaceAllString(url, `api.github.com/repos/`) + "/releases/latest"
}

func ReposListOutput(reposList []string) string {
	if len(reposList) == 0 {
		return "There is no repos at this moment."
	}
	return strings.Join(reposList, "\n")
}

func LatestListOutput(latestList []internal.LatestRelease) string {
	var latest []string

	// Count length of repo name and tag name strings to format output table
	// ToDo: make one loop for count both lenRepo, lenTag
	var lenRepo, lenTag int

	if len(latestList) == 0 {
		return "There is no releases at this moment."
	} else {
		for _, s := range latestList {
			re := regexp.MustCompile(`https://github.com/`)
			repo := re.ReplaceAllString(s.RepoName, "${1}")
			if len(repo) < lenRepo {
				continue
			}
			if len(repo) > lenRepo {
				lenRepo = len(repo)
			}
		}

		for _, s := range latestList {
			if len(s.TagName) < lenTag {
				continue
			}
			if len(s.TagName) > lenTag {
				lenTag = len(s.TagName)
			}
		}
	}

	// Format output table
	latest = append(latest, "<pre>")
	latest = append(latest, fmt.Sprintf("+%s+%s+%s+", strings.Repeat("-", lenRepo+2), strings.Repeat("-", lenTag+2), strings.Repeat("-", 12)))
	latest = append(latest, fmt.Sprintf("| Name%s| Tag%s| Date%s|", strings.Repeat(" ", lenRepo-3), strings.Repeat(" ", lenTag-2), strings.Repeat(" ", 7)))
	latest = append(latest, fmt.Sprintf("+%s+%s+%s+", strings.Repeat("-", lenRepo+2), strings.Repeat("-", lenTag+2), strings.Repeat("-", 12)))
	for _, data := range latestList {
		var tag, date, r string

		re := regexp.MustCompile(`https://github.com/`)
		repo := re.ReplaceAllString(data.RepoName, "${1}")

		if len(repo) < lenRepo {
			repo = fmt.Sprintf("%s%s", repo, strings.Repeat(" ", lenRepo-len(repo)))
		}

		if len(data.TagName) < lenTag {
			tag = fmt.Sprintf("%s%s", data.TagName, strings.Repeat(" ", lenTag-len(data.TagName)))
		} else {
			tag = data.TagName
		}

		date = data.PublishedAt.Format("02.01.2006")

		r = fmt.Sprintf("| %s | %s | %s |", repo, tag, date)
		latest = append(latest, r)
	}
	latest = append(latest, fmt.Sprintf("+%s+%s+%s+", strings.Repeat("-", lenRepo+2), strings.Repeat("-", lenTag+2), strings.Repeat("-", 12)))
	latest = append(latest, "</pre>")
	return strings.Join(latest, "\n")
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
	re, err := regexp.Compile(`https://github.com/`)
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
