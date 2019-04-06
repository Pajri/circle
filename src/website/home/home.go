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

	val := r.FormValue("vote")
	if val != "" {
		fmt.Println("val : ", val)
	}

	//process session
	session, err = utils.GetSession(r, utils.SESSION_AUTH)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "home : an error occured when executing templates")
		return
	}

	//init home
	err = initHome()
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

	vote(r)

	homeViewModel.Questions, err = listQuestion(homeViewModel.CurrentPage)
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

func initHome() error {
	var err error

	ctx := context.TODO()
	db, err = utils.ConnectDb(ctx)
	return err
}

func listQuestion(page int) ([]*datamodel.Question, error) {
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

		question.ID = m[datamodel.FieldQuestionID].(primitive.ObjectID).Hex()
		question.Title = m[question.TitleColl()].(string)
		question.Description = m[question.DescriptionColl()].(string)[:100] + "..."
		question.Vote = int(m[question.VoteColl()].(int32))
		question.IsSolved = m[question.IsSolvedColl()].(bool)
		question.Username = m[question.UsernameColl()].(string)

		fmt.Println("DateTime : ", m[question.CreatedDateColl()].(primitive.DateTime))

		questions = append(questions, question)
	}

	return questions, nil
}

func vote(r *http.Request) error {
	var ok bool
	var val []string
	val, ok = r.URL.Query()["action"]
	if !ok {
		//action not passed, no voting
		return nil
	}
	action := val[0]
	var add int
	if action == "upvote" {
		add = 1
	} else if action == "downvote" {
		add = -1
	}

	//get id from querystring
	val, ok = r.URL.Query()["id"]
	id := val[0]
	objId, _ := primitive.ObjectIDFromHex(id)
	idDoc := bson.D{{"_id", objId}} //object id for filtering

	var q datamodel.Question
	proj := bson.D{{"vote", 1}} //projection : only show id and vote
	err := db.Collection(datamodel.CollQuestion).FindOne(ctx, idDoc, options.FindOne().SetProjection(proj)).Decode(&q)
	if err != nil {
		return err
	}

	updateDoc := bson.D{{datamodel.FieldQuestionVote, q.Vote + add}} //update vote data
	_, err = db.Collection(datamodel.CollQuestion).UpdateOne(ctx, idDoc, bson.D{{"$set", updateDoc}})
	if err != nil {
		return err
	}

	return nil
}
