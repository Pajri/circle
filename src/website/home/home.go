package home

import (
	"context"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/gorilla/sessions"
	"go.mongodb.org/mongo-driver/mongo"

	datamodel "../../data_model"
	utils "../../utils"
)

var funcMap = template.FuncMap{"formatCreatedDate": formatCreatedDate}

//template used in home
var homeTemplates = template.Must(template.New("home.html").Funcs(funcMap).ParseFiles(utils.WebsiteDirectory()+"/home/home.html",
	utils.WebsiteDirectory()+"/layout/main.html"))

var db *mongo.Database
var ctx context.Context

type HomeViewModel struct {
	Questions   []QuestionsView
	CurrentPage int
	PageIndex   []int
}

type QuestionsView struct {
	Question datamodel.Question
	IsVoted  bool
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

	homeViewModel.Questions, err = listQuestion(homeViewModel.CurrentPage, r)
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

func listQuestion(page int, r *http.Request) ([]QuestionsView, error) {
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

	var username string
	username, err = getUsernameFromSession(r)
	if err != nil {
		return nil, err
	}

	var usr datamodel.User
	usr, err = getUser(username)
	if err != nil {
		return nil, err
	}

	var questions []QuestionsView
	for c.Next(ctx) {
		doc := &bson.D{}
		c.Decode(doc)
		m := doc.Map()

		var question datamodel.Question
		question.ID = m[datamodel.FieldQuestionID].(primitive.ObjectID).Hex()
		question.Title = m[datamodel.FieldQuestionTitle].(string)
		question.Description = m[datamodel.FieldQuestionDescription].(string)[:100] + "..."
		question.Vote = int(m[datamodel.FieldQuestionVote].(int32))
		question.IsSolved = m[datamodel.FieldQuestionIsSolved].(bool)
		question.Username = m[datamodel.FieldQuestionUsername].(string)
		question.CreatedDate = utils.UnixTimeToTime(m[datamodel.FieldQuestionCreatedDate].(primitive.DateTime))

		//check if user already voted or not
		var voted = false
		if usr.Vote != nil {
			qObjId := m[datamodel.FieldQuestionID].(primitive.ObjectID)
			voted = isVoted(qObjId, usr.Vote)
		}

		q := QuestionsView{
			Question: question,
			IsVoted:  voted,
		}
		questions = append(questions, q)
	}

	return questions, nil
}

func getUser(username string) (datamodel.User, error) {
	usnameDoc := bson.D{{datamodel.FieldUserUsername, username}}
	usrDoc := bson.D{}
	err := db.Collection(datamodel.CollUser).FindOne(ctx, usnameDoc).Decode(&usrDoc)
	if err != nil {
		return datamodel.User{}, err
	}

	m := usrDoc.Map()
	var usr datamodel.User
	usr.ID = m[datamodel.FieldUserID].(primitive.ObjectID).Hex()
	usr.Username = m[datamodel.FieldUserUsername].(string)
	usr.Email = m[datamodel.FieldUserEmail].(string)

	if m[datamodel.FieldUserVote] != nil {
		usr.Vote = m[datamodel.FieldUserVote].(primitive.A)
	} else {
		usr.Vote = nil
	}
	return usr, nil
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
	questionId := val[0]
	questionObjId, _ := primitive.ObjectIDFromHex(questionId)
	idDoc := bson.D{{"_id", questionObjId}} //object id for filtering

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

	//add voted question to user
	//get username from session
	var username string
	username, err = getUsernameFromSession(r)
	if err != nil {
		return err
	}

	//get user from database
	usnameDoc := bson.D{{datamodel.FieldUserUsername, username}} //username doc for filtering
	usrDoc := bson.D{}
	proj = bson.D{{datamodel.FieldUserVote, 1}} //show only id and vote
	err = db.Collection(datamodel.CollUser).FindOne(ctx, usnameDoc, options.FindOne().SetProjection(proj)).Decode(&usrDoc)
	if err != nil {
		return err
	}

	//create vote array
	m := usrDoc.Map()
	var voteArray bson.A
	if m[datamodel.FieldUserVote] != nil {
		voteArray = m[datamodel.FieldUserVote].(bson.A)
		if !isVoted(questionObjId, voteArray) {
			voteArray = append(voteArray, questionObjId)
		}
	} else {
		voteArray = bson.A{questionObjId}
	}

	//update vote
	votedUpdate := bson.D{{datamodel.FieldQuestionVote, voteArray}}
	_, err = db.Collection(datamodel.CollUser).UpdateOne(ctx, usnameDoc, bson.D{{"$set", votedUpdate}})
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func getUsernameFromSession(r *http.Request) (string, error) {
	var session *sessions.Session
	var err error
	session, err = utils.GetSession(r, utils.SESSION_AUTH)
	if err != nil {
		return "", err
	}
	username := session.Values[utils.KEY_USERNAME].(string)
	return username, nil
}

func isVoted(id primitive.ObjectID, voteArr primitive.A) bool {
	for _, elmt := range voteArr {
		if elmt == id {
			return true
		}
	}
	return false
}

func formatCreatedDate(t time.Time) string {
	y, m, d := t.Date()
	return fmt.Sprintf("%v %v %v", d, m, y)
}
