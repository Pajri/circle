package register

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	datamodel "../../data_model"
	utils "../../utils"
)

var ctx context.Context
var db *mongo.Database

type RegisterView struct {
	IsError      bool
	ErrorMessage string
}

//template used in home
var registerTemplate = template.Must(template.ParseFiles(utils.WebsiteDirectory()+"/register/register.html",
	utils.WebsiteDirectory()+"/layout/authentication.html"))

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	err := Init()
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "register : an error occured when initializing register")
		return
	}

	regView := RegisterView{
		IsError:      false,
		ErrorMessage: "",
	}

	if r.Method == "POST" {
		var taken, passSame bool
		taken, err = isUserTaken(r.FormValue("username"))
		passSame = isPasswordSame(r)

		if !taken && passSame {
			err = register(w, r)
			if err == nil {
				http.Redirect(w, r, "/login?register=success", 302)
			}else{
				utils.InternalServerErrorHandler(w, r, err, "register : an error occured when registering user")
				return
			}
		} else {
			if taken {
				regView.IsError = true
				regView.ErrorMessage = "Username is already taken"
			} else if !passSame {
				regView.IsError = true
				regView.ErrorMessage = "Password does not match"
			}
		}
	}

	err = registerTemplate.ExecuteTemplate(w, "authentication.html", regView)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "register : an error occured when executing template")
		return
	}

}

func Init() error {
	//INIT CONTEXT
	//TODO need to understand what ctx is and how to use it
	ctx = context.TODO()

	//INIT DATABASE
	var err error
	db, err = utils.ConnectDb(ctx)
	return err
}

func register(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return nil
	}

	err := insertUser(r)
	return err
}

func isUserTaken(username string) (bool, error) {
	findOptions := options.Find()
	findOptions = findOptions.SetLimit(1)
	userFound, err := db.Collection(datamodel.CollUser).Find(ctx, bson.D{{datamodel.FieldUserUsername, username}}, findOptions)

	if err != nil {
		return false, err
	}

	return userFound.Next(ctx), nil
}

func isPasswordSame(r *http.Request) bool {
	if r.FormValue("password") == r.FormValue("confirm_password") {
		return true
	}
	return false
}

func insertUser(r *http.Request) error {
	_, err := db.Collection(datamodel.CollUser).InsertOne(ctx, bson.D{
		{datamodel.FieldUserUsername, r.FormValue("username")},
		{datamodel.FieldUserEmail, r.FormValue("email"),
		{datamodel.FieldPassword, utils.HashSha1(r.FormValue("password")},
	})

	return err
}
