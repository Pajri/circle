package home

import (
	"strings"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"fmt"
	"context"
	"html/template"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/mongo"

	datamodel "../../data_model"
	utils "../../utils"
)

//template used in home
var homeTemplates = template.Must(template.ParseFiles(utils.WebsiteDirectory()+"/home/home.html",
	utils.WebsiteDirectory()+"/layout/main.html"))

var db *mongo.Database
var ctx context.Context

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	splitString := strings.Split(r.URL.Path,"/")
	fmt.Println("split string : ",splitString)
	if len(splitString) > 2 {
		fmt.Println("page number :",splitString[3])
	}
	
	var err error
	var session *sessions.Session

	session, err = utils.GetSession(r, utils.SESSION_AUTH)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "home : an error occured when executing templates")
		return
	}

	err = InitHome()
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "home : an error occured when executing templates")
		return
	}

	isAuth := utils.IsLoggedInSession(session)
	if !isAuth {
		utils.ForbiddenHandler(w, r)
		return
	}

	var questions []*datamodel.Question
	questions, err = ListQuestion()
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "home : an error occured when load list question")
		return
	}

	err = homeTemplates.ExecuteTemplate(w, "main.html", questions)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "home : an error occured when executing templates")
		return
	}
}

func InitHome() error {
	var err error

	ctx := context.TODO()
	db, err = utils.ConnectDb(ctx)
	return err
}

func ListQuestion() ([]*datamodel.Question, error) {
	opt := options.Find()
	opt.SetLimit(5)
	opt.SetSort(bson.D{{datamodel.FieldQuestionCreatedDate, -1}})
	
	c, err := db.Collection(datamodel.CollQuestion).Find(ctx, bson.D{},opt)
	if err != nil {
		return nil, err
	}
	defer c.Close(ctx)

	var questions []*datamodel.Question
	for c.Next(ctx) {
		question := new(datamodel.Question)
		// err = c.Decode(&question)
		// if err != nil {
		// 	return nil, err
		// }

		doc := &bson.D{}
		c.Decode(doc)
		m := doc.Map()

		question.Title = m[question.TitleColl()].(string)
		question.Description = m[question.DescriptionColl()].(string)[:100]+"..."
		question.Vote = int(m[question.VoteColl()].(float64))
		question.IsSolved = m[question.IsSolvedColl()].(bool)
		question.Username = m[question.UsernameColl()].(string)

		fmt.Println("DateTime : ", m[question.CreatedDateColl()].(primitive.DateTime))

		questions = append(questions, question)
	}

	return questions, nil
}
