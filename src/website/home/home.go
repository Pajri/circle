package home

import (
	"context"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

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

type HomeViewModel struct {
	Questions   []*datamodel.Question
	CurrentPage int
	PageIndex   []int
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var session *sessions.Session
	homeViewModel := HomeViewModel{
		Questions:   nil,
		CurrentPage: 1,
		PageIndex:   nil,
	}

	//process session
	session, err = utils.GetSession(r, utils.SESSION_AUTH)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "home : an error occured when executing templates")
		return
	}

	//init home
	err = InitHome()
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "home : an error occured when executing templates")
		return
	}

	//check is logged in
	isAuth := utils.IsLoggedInSession(session)
	if !isAuth {
		utils.ForbiddenHandler(w, r)
		return
	}

	//get page for pagination
	splitUrl := strings.Split(r.URL.Path, "/")
	if len(splitUrl) == 4 { // the url like /home/page/3, the first element is empty string : [,home,page,3]
		if splitUrl[2] == "page" {
			var cur int64
			cur, err = strconv.ParseInt(splitUrl[3], 10, 64)
			if err != nil {
				utils.InternalServerErrorHandler(w, r, err, "home : an error occured when casting page num")
				return
			}
			homeViewModel.CurrentPage = int(cur)
		}
	}

	count := int64(0)
	count, err = db.Collection(datamodel.CollQuestion).CountDocuments(ctx, bson.D{}, nil)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "home : an error occured when counting documents")
		return
	}
	pagesCount := int(math.Ceil(float64(count) / 5))
	homeViewModel.PageIndex = make([]int, pagesCount)
	for i := 0; i < int(pagesCount); i++ {
		homeViewModel.PageIndex[i] = i + 1
	}

	homeViewModel.Questions, err = ListQuestion(homeViewModel.CurrentPage)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "home : an error occured when load list question")
		return
	}

	err = homeTemplates.ExecuteTemplate(w, "main.html", homeViewModel)
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

func ListQuestion(page int) ([]*datamodel.Question, error) {
	if page > 0 {
		page = page - 1
	}

	opt := options.Find()
	limit := 5
	opt.SetLimit(int64(limit))
	opt.SetSort(bson.D{{datamodel.FieldQuestionCreatedDate, -1}})
	opt.SetSkip(int64(page * limit))

	c, err := db.Collection(datamodel.CollQuestion).Find(ctx, bson.D{}, opt)
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
		question.Description = m[question.DescriptionColl()].(string)[:100] + "..."
		question.Vote = int(m[question.VoteColl()].(float64))
		question.IsSolved = m[question.IsSolvedColl()].(bool)
		question.Username = m[question.UsernameColl()].(string)

		fmt.Println("DateTime : ", m[question.CreatedDateColl()].(primitive.DateTime))

		questions = append(questions, question)
	}

	return questions, nil
}
