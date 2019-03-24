package utils

import (
	"html/template"
	"fmt"
	"net/http"
)

func InternalServerErrorHandler(w http.ResponseWriter, r *http.Request, err error, msg string){
	fmt.Println(fmt.Errorf(msg))
	fmt.Println(fmt.Errorf("%v",err))

	errTmpl := template.Must(template.ParseFiles(WebsiteDirectory()+"/error_pages/500.html"))
	errTmpl.Execute(w, nil)
}