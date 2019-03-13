package home

import (
	"html/template"
	"net/http"

	utils "../../utils"
)

//template used in home
var homeTemplates = template.Must(template.ParseFiles(utils.WebsiteDirectory()+"/home/home.html",
	utils.WebsiteDirectory()+"/layout/main.html"))

func LoadHome(w http.ResponseWriter) error {
	err := homeTemplates.ExecuteTemplate(w, "main.html", nil)
	if err != nil {
		return err
	}
	return nil
}
