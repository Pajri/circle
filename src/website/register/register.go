package register

import (
	"context"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"

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
	err := initRegister()
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
			} else {
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

	db.Client().Disconnect(ctx)
}

func initRegister() error {
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
	img, err := uploadImage(r)
	if err != nil {
		return err
	}

	_, err = db.Collection(datamodel.CollUser).InsertOne(ctx, bson.D{
		{datamodel.FieldUserUsername, r.FormValue("username")},
		{datamodel.FieldUserEmail, r.FormValue("email")},
		{datamodel.FieldPassword, utils.HashSha1(r.FormValue("password"))},
		{datamodel.FieldUserImageName, img},
	})

	return err
}

func uploadImage(r *http.Request) (string, error) {
	f, h, err := r.FormFile("image") //get submitted image
	if err != nil {
		return "", err
	}
	defer f.Close()

	var imgFile *os.File
	usname := r.FormValue("username")
	filename := usname + filepath.Ext(h.Filename)                                       //create filename
	imgFile, err = os.Create(utils.WorkingDirectory() + "/upload/userdata/" + filename) //upload file
	if err != nil {
		return "", err
	}
	defer imgFile.Close()

	_, err = io.Copy(imgFile, f)
	if err != nil {
		return "", err
	}
	return filename, nil
}
