package ask

import (
	"context"
	"html/template"
	"net/http"
	"time"

	"github.com/gorilla/sessions"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/mongo"

	datamodel "../../data_model"
	utils "../../utils"
)

var askTemplate = template.Must(template.ParseFiles(utils.WebsiteDirectory()+"/ask/ask.html",
	utils.WebsiteDirectory()+"/layout/main.html"))

var ctx context.Context
var db *mongo.Database
var s *sessions.Session

func AskHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	err = initAsk(r)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "ask : an error occured while initializing ask")
		return

	}

	if r.Method == "POST" {
		var qDoc primitive.D
		qDoc = createQuestionDoc(r)

		err = insertQuestion(qDoc)
		if err != nil {
			utils.InternalServerErrorHandler(w, r, err, "ask : an error occured while inserting document")
			return
		} else {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		}
	}

	err = askTemplate.ExecuteTemplate(w, "main.html", nil)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "ask : an error occured when executing template")
		return
	}
}

func initAsk(r *http.Request) error {
	var err error
	ctx := context.TODO()          //init context
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

func createQuestionDoc(r *http.Request) bson.D {
	t := r.FormValue("title")
	qTxt := r.FormValue("question")

	var usname string
	usname = utils.GetUsernameFromSession(s, r)
	qDoc := bson.D{
		{datamodel.FieldQuestionID, primitive.NewObjectID()},
		{datamodel.FieldQuestionTitle, t},
		{datamodel.FieldQuestionDescription, qTxt},
		{datamodel.FieldQuestionVote, 0},
		{datamodel.FieldQuestionIsSolved, false},
		{datamodel.FieldQuestionUsername, usname},
		{datamodel.FieldQuestionCreatedDate, primitive.DateTime(utils.TimeToMillis(time.Now()))},
	}

	return qDoc
}

func insertQuestion(qDoc bson.D) error {
	_, err := db.Collection(datamodel.CollQuestion).InsertOne(ctx, qDoc)
	return err
}
