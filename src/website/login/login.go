package login

import (
	"go.mongodb.org/mongo-driver/mongo"
	"context"
	"html/template"
	"net/http"
	
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"

	datamodel "../../data_model"
	utils "../../utils"
)

var ctx context.Context
var db utils.DbUtil

var loginTemplate = template.Must(template.ParseFiles(utils.WebsiteDirectory()+"/login/login.html",
	utils.WebsiteDirectory()+"/layout/authentication.html"))

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	initHandler()

	if r.Method == "POST" {
		usrLogin := populateUsernamePassword(r)

	}
	_ = loginTemplate.ExecuteTemplate(w, "authentication.html", nil)

	db.Disconnect()
}

func initHandler(){
	ctx = context.TODO()
	db := new(utils.DbUtil)
	db.Connect(ctx)
}

func authenticate(usrLogin datamodel.User) (bool, error) {
	user := new(datamodel.User)
	findOptions := options.Find()
	findOptions = findOptions.SetLimit(1)
	userFound, err := db.Collection(user.CollName()).Find(ctx, bson.D{{user.UsernameColl(), username}}, findOptions)
	return false, nil
}

func populateUsernamePassword(r *http.Request) datamodel.User {
	return datamodel.User{
		Username: r.FormValue("username"),
		Password: utils.HashSha1(r.FormValue("password")),
	}
}
