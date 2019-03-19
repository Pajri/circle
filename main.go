package main

import (
	"log"
	"net/http"

	discussion "./src/website/discussion"
	home "./src/website/home"
	register "./src/website/register"
	login "./src/website/login"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	home.LoadHome(w)
}

func discussionHandler(w http.ResponseWriter, r *http.Request) {
	discussion.LoadDiscussion(w)
}

func main() {
	assetsDir := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", assetsDir))

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/register", register.RegisterHandler)
	http.HandleFunc("/discussion", discussionHandler)
	http.HandleFunc("/login", login.LoginHandler)


	log.Fatal(http.ListenAndServe(":8080", nil))
}
