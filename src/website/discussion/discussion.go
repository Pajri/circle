package discussion

import (
	"net/http"
	"text/template"

	utils "../../utils"
)

var discussionTemplate = template.Must(template.ParseFiles(utils.WebsiteDirectory() + "/discussion/discussion.html"))

func LoadDiscussion(w http.ResponseWriter) error {
	err := discussionTemplate.ExecuteTemplate(w, "discussion.html", nil)
	if err != nil {
		return err
	}
	return nil
}
