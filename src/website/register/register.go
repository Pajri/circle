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

type RegisterViewModel struct {
	IsError      bool
	ErrorMessage string
}

//template used in home
var registerTemplate = template.Must(template.ParseFiles(utils.WebsiteDirectory()+"/register/register.html",
	utils.WebsiteDirectory()+"/layout/authentication.html"))

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	err := Init()

	RegisterData := RegisterViewModel{
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
			}
		} else {
			if taken {
				RegisterData.IsError = true
				RegisterData.ErrorMessage = "Username is already taken"
			} else if !passSame {
				RegisterData.IsError = true
				RegisterData.ErrorMessage = "Password does not match"
			}
		}
	}

	err = registerTemplate.ExecuteTemplate(w, "authentication.html", RegisterData)

	if err != nil {
		log.Print("Register : ", err)
	}

}

func Init() error {
	//INIT CONTEXT
	//TODO need to understand what ctx is and how to use it
	ctx = context.TODO()

	//INIT DATABASE
	dbInit, err := utils.ConfigDB(ctx)
	if err != nil {
		return fmt.Errorf("An error occured on config db: %v", err)
	}
	db = dbInit
	return nil
}

func register(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return nil
	}

	var err error
	user := populateUser(r)
	if err != nil {
		return err
	}

	err = insertUser(user)
	if err != nil {
		return fmt.Errorf("User couldn't be created: %v", err)
	}
	return err

}

func isUserTaken(username string) (bool, error) {
	user := new(datamodel.User)
	findOptions := options.Find()
	findOptions = findOptions.SetLimit(1)
	userFound, err := db.Collection(user.CollName()).Find(ctx, bson.D{{user.UsernameColl(), username}}, findOptions)

	if err != nil {
		return false, fmt.Errorf("An error occured while checking user taken: %v", err)
	}

	return userFound.Next(ctx), nil
}

func isPasswordSame(r *http.Request) bool {
	if r.FormValue("password") == r.FormValue("confirm_password") {
		return true
	}
	return false
}

func insertUser(user datamodel.User) error {
	_, err := db.Collection(user.CollName()).InsertOne(ctx, bson.D{
		{user.UsernameColl(), user.Username},
		{user.EmailColl(), user.Email},
		{user.PasswordColl(), user.Password},
	})

	return err
}

func populateUser(r *http.Request) datamodel.User {
	user := datamodel.User{
		Username: r.FormValue("username"),
		Email:    r.FormValue("email"),
		Password: utils.HashSha1(r.FormValue("password")),
	}

	return user
}
