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

	// Count length of repo name and tag name strings to format output
	lenRepo := -1
	lenTag := -1
	for _, s := range latestList {
		re := regexp.MustCompile(`https://github.com/`)
		repo := re.ReplaceAllString(s.RepoName, "${1}")
		if len(repo) < lenRepo {
			continue
		}
		if len(repo) > lenRepo {
			lenRepo = len(repo)
		}
		if len(s.TagName) < lenTag {
			continue
		}
		if len(s.TagName) > lenTag {
			lenTag = len(s.TagName)
		}
	}

	log.Printf("DEBUG: repoLen %v\n", lenRepo)
	log.Printf("DEBUG: tagLen %v\n", lenTag)

	if len(latestList) == 0 {
		return "There is no releases at this moment."
	}

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

		//r = fmt.Sprintf("%-20s | %-20s | %-20s |\n", data.RepoName, data.TagName, data.PublishedAt.Format("02.01.2006 15:04:05"))
		r = fmt.Sprintf("| %s | %s | %s |", repo, tag, date)
		latest = append(latest, r)
		//latest = append(latest, data.RepoName+" "+"<b>"+data.TagName+"</b> released at: "+data.PublishedAt.Format("02.01.2006 15:04:05"))
	}
	latest = append(latest, fmt.Sprintf("+%s+%s+%s+", strings.Repeat("-", lenRepo+2), strings.Repeat("-", lenTag+2), strings.Repeat("-", 12)))
	latest = append(latest, "</pre>")
	return strings.Join(latest, "\n")
}

//func LatestListOutputV2(latestList []internal.LatestRelease) string {
//	var latest []string
//
//	if len(latestList) == 0 {
//		return "There is no releases at this moment."
//	}
//	// Count length of repo name and tag name strings to format output
//	//lenRepo := -1
//	//lenTag := -1
//	//for _, s := range latestList {
//	//	if len(s.RepoName) < lenRepo {
//	//		continue
//	//	}
//	//	if len(s.RepoName) > lenRepo {
//	//		lenRepo = len(s.RepoName)
//	//	}
//	//	if len(s.TagName) < lenTag {
//	//		continue
//	//	}
//	//	if len(s.TagName) > lenTag {
//	//		lenTag = len(s.RepoName)
//	//	}
//	//}
//	//log.Printf("DEBUG: lenRepo %v\n", lenRepo)
//	//log.Printf("DEBUG: lenTag %v\n", lenTag)
//
//	//for _, data := range latestList {
//	//	var r string
//	//	r = fmt.Sprintf("%-20s | %-20s | %-20s |\n", data.RepoName, data.TagName, data.PublishedAt.Format("02.01.2006 15:04:05"))
//	//	latest = append(latest, r)
//	//	//latest = append(latest, data.RepoName+" "+"<b>"+data.TagName+"</b> released at: "+data.PublishedAt.Format("02.01.2006 15:04:05"))
//	//	log.Printf("DEBUG: %v", r)
//	//}
//	//return strings.Join(latest, "\n")
//
//	columnLengths := make([]int, 3)
//	lenRepo := -1
//	lenTag := -1
//	for _, s := range latestList {
//		if len(s.RepoName) < lenRepo {
//			continue
//		}
//		if len(s.RepoName) > lenRepo {
//			lenRepo = len(s.RepoName)
//		}
//		if len(s.TagName) < lenTag {
//			continue
//		}
//		if len(s.TagName) > lenTag {
//			lenTag = len(s.RepoName)
//		}
//		//for i, val := range line {
//		//	if len(val[i]) > columnLengths[i] {
//		//		columnLengths[i] = len(val)
//		//	}
//		//}
//	}
//
//	columnLengths = append(columnLengths, lenRepo)
//	columnLengths = append(columnLengths, lenTag)
//	columnLengths = append(columnLengths, 19)
//
//	fmt.Printf("DEBUG:columnLengths %v\n", columnLengths)
//
//	var lineLength int
//	for _, c := range columnLengths {
//		lineLength += c + 3 // +3 for 3 additional characters before and after each field: "| %s "
//	}
//	lineLength += 1
//
//	for i, line := range latestList {
//		if i == 0 { // table header
//			//var s string
//			//s = fmt.Sprintf("+%s+\n", strings.Repeat("-", lineLength-2))
//			latest = append(latest, fmt.Sprintf("+%s+\n", strings.Repeat("-", lineLength-2))) // lineLength-2 because of "+" as first and last character
//		}
//		for j, val := range line {
//			latest = append(latest, fmt.Sprintf("| %-*s |\n", columnLengths[j], val))
//			//if j == len(line)-1 {
//			//	latest = append(latest, fmt.Sprintf("|\n"))
//			//}
//		}
//		if i == 0 || i == len(latestList)-1 { // table header or last line
//			fmt.Printf("+%s+\n", strings.Repeat("-", lineLength-2)) // lineLength-2 because of "+" as first and last character
//		}
//	}
//	return strings.Join(latest, "\n")
//}

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
