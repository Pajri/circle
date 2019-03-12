package register

import (
	"html/template"
	"net/http"

	utils "../../utils"
)

//template used in home
var registerTemplate = template.Must(template.ParseFiles(utils.WebsiteDirectory() + "/register/register.html"))

func LoadRegister(w http.ResponseWriter) error {
	err := registerTemplate.ExecuteTemplate(w, "register.html", nil)
	if err != nil {
		return err
	}
	return nil
}
