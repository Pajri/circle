package utils

import (
	"crypto/sha1"
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

func HashSha1(text string) string {
	h := sha1.New()
	h.Write(([]byte(text)))
	return string(h.Sum(nil))
}
