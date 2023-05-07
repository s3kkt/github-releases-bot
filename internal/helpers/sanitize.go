package helpers

import (
	"errors"
	"log"
	"reflect"
	"strconv"
	"strings"
)

func ReposListOutput(reposList []string) string {
	return strings.Join(reposList, "\n")
}

func GetArgFromCommand(command string) (string, error) {
	s := strings.Fields(command)
	count := len(s) - 1
	if count == 1 {
		return s[1], nil
	} else if len(s) == 1 {
		err := errors.New("command takes at least one argument")
		return "", err
	} else {
		err := errors.New("must be only one argument, got " + strconv.Itoa(count))
		return "", err
	}
}

func ValidateArgIsString(arg string) (bool, error) {
	if reflect.TypeOf(arg).Kind() == reflect.String && arg != "" {
		return true, nil
	} else {
		err := errors.New("Argument must be a string type, got " + string(reflect.TypeOf(arg).Kind()))
		return false, err
	}
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
