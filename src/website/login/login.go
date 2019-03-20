package login

import (
	"fmt"
	"html/template"
	"net/http"
	utils "../../utils"
)

//TODO check if this can be separated to make it reusasble
type Message struct {
	Display bool
	Text string
	Type string
}

var loginTemplate = template.Must(template.ParseFiles(utils.WebsiteDirectory()+"/login/login.html",
	utils.WebsiteDirectory()+"/layout/authentication.html"))

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	LoginMessage := Message{
		Display : false,
		Text : "",
		Type : "",
	}

	param, _ := r.URL.Query()["register"]
	if param[0] == "success" {
		LoginMessage = createSuccessMessage("Registration successful")
		err := loginTemplate.ExecuteTemplate(w,"authentication.html", LoginMessage)
		if err != nil {
			fmt.Println(err)
		}
	}else{
		_ = loginTemplate.ExecuteTemplate(w, "authentication.html", nil)
	}
	
}

func createSuccessMessage(message string) Message{
	return Message{
		Display : true,
		Text : message,
		Type : "success",
	}

}