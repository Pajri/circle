package login

import (
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"context"
	"html/template"
	"net/http"

	datamodel "../../data_model"
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
var db *mongo.Database
var ctx context.Context

func initLogin() error{
	var err error
	
	ctx := context.TODO()
	db, err = utils.ConnectDb(ctx)
	return err
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	err := initLogin()
	if err != nil {
		utils.InternalServerErrorHandler(w,r,err,"login : an error occured when init login.")
		return
	}

	LoginMessage := Message{
		Display : false,
		Text : "",
		Type : "",
	}

	param, ok := r.URL.Query()["register"]
	if ok && param[0] == "success" {
		LoginMessage = createSuccessMessage("Registration successful")
	}

	if r.Method == "POST" {
		//login process
		login := populateLogin(r)
		isAuth,err := authenticate(login)
		if err != nil {
			utils.InternalServerErrorHandler(w,r,err,"login : an error occured when authenticating.")
			return
		}

		if isAuth {
			http.Redirect(w,r,"/home",http.StatusTemporaryRedirect)
			session, err := utils.GetSession(r,utils.SESSION_AUTH)
			if err != nil {
				utils.InternalServerErrorHandler(w,r,err,"login : an error occured when retrieving sessions.")
				return
			}

			session.Values[utils.KEY_USERNAME] = login.Username
			session.Values[utils.KEY_ISAUTH] = true
			err = session.Save(r,w)
			if err != nil {
				utils.InternalServerErrorHandler(w,r,err,"login : an error occured when executing template.")
			}
		}else{
			LoginMessage = createErrorMessage("Invalid username or password")
		}
	}
	
	err = loginTemplate.ExecuteTemplate(w, "authentication.html", LoginMessage)
	if err != nil {
		utils.InternalServerErrorHandler(w,r,err,"login : an error occured when executing template.")
	}

	db.Client().Disconnect(ctx)
}

func populateLogin(r *http.Request) datamodel.User {
	return datamodel.User{
		Username : r.FormValue("username"),
		Password : utils.HashSha1(r.FormValue("password")),
	}
}

func authenticate(login datamodel.User) (bool, error) {
	bsonLogin := bson.D{
		{login.UsernameColl(), login.Username},
		{login.PasswordColl(), login.Password},
	}

	//get single document
	opt := options.Find()
	opt = opt.SetLimit(1)
	usrCur, err := db.Collection(login.CollName()).Find(ctx, bsonLogin, opt) //get cursor and error

	if err != nil {
		return false, err
	}

	isAuth := usrCur.Next(ctx) //return false if there's no document
	return isAuth, nil		

}

func createSuccessMessage(msg string) Message{
	return Message{
		Display : true,
		Text : msg,
		Type : "success",
	}
}

func createErrorMessage(msg string) Message{
	return Message{
		Display : true,
		Text : msg,
		Type : "error",
	}
}