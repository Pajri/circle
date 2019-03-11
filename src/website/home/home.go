package home

import (
	"net/http"
)

import (
	"html/template"
)
//template used in home
var homeTemplates = template.Must(template.ParseFiles("home.html"))

func LoadHome(w http.ResponseWriter) error{
	err := homeTemplates.ExecuteTemplate(w, "home.html", nil)
	if err != nil{
		return err
	}
	return nil
}