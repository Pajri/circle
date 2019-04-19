package main

import (
	"log"
	"net/http"

	ask "./src/website/ask"
	discussion "./src/website/discussion"
	home "./src/website/home"
	login "./src/website/login"
	logout "./src/website/logout"
	profile "./src/website/profile"
	register "./src/website/register"
)

func main() {
	//handle static asset files
	assetsDir := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", assetsDir))

	//handle uploaded files
	fileUpdload := http.FileServer(http.Dir("upload"))
	http.Handle("/upload/", http.StripPrefix("/upload/", fileUpdload))

	http.HandleFunc("/", home.HomeHandler)
	http.HandleFunc("/register", register.RegisterHandler)
	http.HandleFunc("/discussion/", discussion.DiscussionHandler)
	http.HandleFunc("/login", login.LoginHandler)
	http.HandleFunc("/ask", ask.AskHandler)
	http.HandleFunc("/profile/", profile.ProfileHandler)
	http.HandleFunc("/logout", logout.LogoutHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
