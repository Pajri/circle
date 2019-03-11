package main

import (
	"log"
	"net/http"
	home "./website/home"
)

func homeHandler(w http.ResponseWriter, r *http.Request){
	home.LoadHome(w)
}

func main(){
	http.HandleFunc("/", homeHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}