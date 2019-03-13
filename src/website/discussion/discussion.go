package discussion

import (
	"net/http"
	"text/template"

	utils "../../utils"
)

var discussionTemplate = template.Must(template.ParseFiles(utils.WebsiteDirectory()+"/discussion/discussion.html",
	utils.WebsiteDirectory()+"/layout/main.html"))

func LoadDiscussion(w http.ResponseWriter) error {
	err := discussionTemplate.ExecuteTemplate(w, "main.html", nil)
	if err != nil {
		return err
	}
	return nil
}
