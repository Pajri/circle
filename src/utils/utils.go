package utils

import (
	"log"
	"os"
	"strings"
)

func WorkingDirectory() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		return ""
	} else {
		wd = strings.Replace(wd, "\\", "/", -1)
		return wd
	}
}

func WebsiteDirectory() string {
	return WorkingDirectory() + "/src/website"
}
