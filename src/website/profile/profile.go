package profile

import (
	"context"
	"net/http"
	"strings"
	"text/template"

	"github.com/gorilla/sessions"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/bson"

	datamodel "../../data_model"
	utils "../../utils"
	"go.mongodb.org/mongo-driver/mongo"
)

var profileTemplate = template.Must(template.ParseFiles(utils.WebsiteDirectory()+"/profile/profile.html",
	utils.WebsiteDirectory()+"/layout/main.html"))
var ctx context.Context
var db *mongo.Database
var s *sessions.Session

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	err := initProfile(r)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "profile : an error occured when initializing profile")
		return
	}

	var usName string
	usName = getUsernameFromUrl(r)
	if usName == "" {
		usName = utils.GetUsernameFromSession(s, r)
	}

	var usr datamodel.User
	usr, err = getUser(usName)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "profile : an error occured when getting user")
		return
	}

	usr, err = getUser(usName)
	err = profileTemplate.ExecuteTemplate(w, "main.html", usr)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "profile : an error occured when executing templates")
		return
	}

	db.Client().Disconnect(ctx)
}

func initProfile(r *http.Request) error {
	var err error
	ctx = context.TODO()           //init context
	db, err = utils.ConnectDb(ctx) //init db
	if err != nil {
		return err
	}

	s, err = utils.GetSession(r, utils.SESSION_AUTH) //init session
	if err != nil {
		return err
	}

	return err
}

func getUsernameFromUrl(r *http.Request) string {
	split := strings.Split(r.URL.Path, "/")
	if len(split) != 3 {
		return ""
	}
	return split[2]
}

func getUser(username string) (datamodel.User, error) {
	findDoc := bson.D{{datamodel.FieldUserUsername, username}}
	usrDoc := bson.D{}
	err := db.Collection(datamodel.CollUser).FindOne(ctx, findDoc).Decode(&usrDoc)
	if err != nil {
		return datamodel.User{}, err
	}

	usrMap := usrDoc.Map()
	var usr datamodel.User
	usr.ID = usrMap[datamodel.FieldUserID].(primitive.ObjectID).Hex()
	usr.Username = usrMap[datamodel.FieldUserUsername].(string)
	usr.Email = usrMap[datamodel.FieldUserEmail].(string)

	return usr, nil
}
