package home

import (
	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/mongo"
	"html/template"
	"net/http"

	utils "../../utils"
	datamodel "../../data_model"
)

//template used in home
var homeTemplates = template.Must(template.ParseFiles(utils.WebsiteDirectory()+"/home/home.html",
	utils.WebsiteDirectory()+"/layout/main.html"))

var db *mongo.Database
var ctx context.Context

func HomeHandler(w http.ResponseWriter, r *http.Request){
	var err error
	var session *sessions.Session

	session, err = utils.GetSession(r, utils.SESSION_AUTH)
	if err != nil {
		utils.InternalServerErrorHandler(w,r,err,"home : an error occured when executing templates")
		return
	}

	err = InitHome()
	if err != nil {
		utils.InternalServerErrorHandler(w,r,err,"home : an error occured when executing templates")
		return
	}
	
	isAuth := utils.IsLoggedInSession(session)
	if !isAuth{
		utils.ForbiddenHandler(w,r)
		return
	}

	err = homeTemplates.ExecuteTemplate(w, "main.html", nil)
	if err != nil {
		utils.InternalServerErrorHandler(w,r,err,"home : an error occured when executing templates")
		return
	}
}

func InitHome(){
	var err error
	
	ctx := context.TODO()
	db, err = utils.ConnectDb(ctx)
	return err
}

func ListQuestion(){
	c, err := 	db.Collection(new())
}