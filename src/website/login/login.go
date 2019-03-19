package login

import (
	"html/template"
	"net/http"
	utils "../../utils"
)

var loginTemplate = template.Must(template.ParseFiles(utils.WebsiteDirectory()+"/login/login.html",
	utils.WebsiteDirectory()+"/layout/authentication.html"))

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	_ = loginTemplate.ExecuteTemplate(w, "authentication.html", nil)
}