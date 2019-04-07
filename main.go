package main

import (
	"log"
	"net/http"

	discussion "./src/website/discussion"
	home "./src/website/home"
	login "./src/website/login"
	register "./src/website/register"
)

func main() {
	assetsDir := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", assetsDir))

	http.HandleFunc("/", home.HomeHandler)
	http.HandleFunc("/register", register.RegisterHandler)
	http.HandleFunc("/discussion/", discussion.DiscussionHandler)
	http.HandleFunc("/login", login.LoginHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
