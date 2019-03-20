package login

import (
	"html/template"
	"net/http"

	datamodel "../../data_model"
	utils "../../utils"
)

var loginTemplate = template.Must(template.ParseFiles(utils.WebsiteDirectory()+"/login/login.html",
	utils.WebsiteDirectory()+"/layout/authentication.html"))

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		usrLogin := populateUsernamePassword(r)

	}
	_ = loginTemplate.ExecuteTemplate(w, "authentication.html", nil)

}

func authenticate(usrLogin datamodel.User) bool, error {

	return false, nil
}

func populateUsernamePassword(r *http.Request) datamodel.User {
	return datamodel.User{
		Username: r.FormValue("username"),
		Password: utils.HashSha1(r.FormValue("password")),
	}
}
