package register

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	datamodel "../../data_model"
	utils "../../utils"
)

//template used in home
var registerTemplate = template.Must(template.ParseFiles(utils.WebsiteDirectory()+"/register/register.html",
	utils.WebsiteDirectory()+"/layout/authentication.html"))

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		db, err := utils.ConfigDB(ctx)
		if err != nil {
			fmt.Errorf("Register: en error occured on config db: %v", err)
		}

		user := datamodel.User{
			Username: r.FormValue("username"),
			Email:    r.FormValue("email"),
			Password: r.FormValue("password"),
		}
		err = register(ctx, db, user)
		if err != nil {
			fmt.Errorf("Register: User couldn't be created: %v", err)
		}
	}

	err := registerTemplate.ExecuteTemplate(w, "authentication.html", nil)
	if err != nil {
		//TODO handle error
	}
}

func register(ctx context.Context, db *mongo.Database, user datamodel.User) error {

	_, err := db.Collection(user.CollName()).InsertOne(ctx, bson.D{
		{user.UsernameColl(), user.Username},
		{user.EmailColl(), user.Email},
		{user.PasswordColl(), user.Password},
	})

	return err

}
