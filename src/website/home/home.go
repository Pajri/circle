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

var funcMap = template.FuncMap{"formatCreatedDate": utils.FormatCreatedDate}

//template used in home
var homeTemplates = template.Must(template.New("home.html").Funcs(funcMap).ParseFiles(utils.WebsiteDirectory()+"/home/home.html",
	utils.WebsiteDirectory()+"/layout/main.html"))

var db *mongo.Database
var ctx context.Context
var s *sessions.Session

type HomeView struct {
	Questions   []QuestionsView
	CurrentPage int
	PageIndex   []int
}

type QuestionsView struct {
	Question        datamodel.Question
	IsVoted         bool
	NumberOfAnswers int
	IsSolved        bool
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	homeView := HomeView{
		Questions:   nil,
		CurrentPage: 1,
		PageIndex:   nil,
	}

	//init home
	err = initHome(r)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "home : an error occured when executing templates")
		return
	}

	//check is logged in
	isAuth := utils.IsLoggedInSession(s)
	if !isAuth {
		utils.ForbiddenHandler(w, r)
		return
	}

	homeView.CurrentPage, homeView.PageIndex, err = createPagination(r)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "home : an error occured while crating pagination")
	}

	vote(r)

	homeView.Questions, err = listQuestion(homeView.CurrentPage, r)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "home : an error occured when load list question")
		return
	}

	err = homeTemplates.ExecuteTemplate(w, "main.html", homeView)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "home : an error occured when executing templates")
		return
	}

	db.Client().Disconnect(ctx)
}

func initHome(r *http.Request) error {
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
	return nil
}

func listQuestion(page int, r *http.Request) ([]QuestionsView, error) {
	if page > 0 {
		page = page - 1
	}

	//limit only 5 item per page
	opt := options.Find()
	limit := 5
	opt.SetLimit(int64(limit))
	opt.SetSort(bson.D{{datamodel.FieldQuestionCreatedDate, -1}}) //sort by date descending
	opt.SetSkip(int64(page * limit))

	c, err := db.Collection(datamodel.CollQuestion).Find(ctx, bson.D{}, opt)
	if err != nil {
		return nil, err
	}
	defer c.Close(ctx)

	username := utils.GetUsernameFromSession(s, r)

	var usr datamodel.User
	usr, err = getUser(username)
	if err != nil {
		return nil, err
	}

	var questions []QuestionsView
	for c.Next(ctx) {
		//get question
		doc := &bson.D{}
		c.Decode(doc)
		m := doc.Map()
		var question datamodel.Question
		question.ID = m[datamodel.FieldQuestionID].(primitive.ObjectID).Hex()
		question.Title = m[datamodel.FieldQuestionTitle].(string)
		question.Description = m[datamodel.FieldQuestionDescription].(string)
		if len(question.Description) > 100 {
			question.Description = question.Description[:100] + "..."
		}

		question.Vote = m[datamodel.FieldQuestionVote].(int32)
		question.IsSolved = m[datamodel.FieldQuestionIsSolved].(bool)
		question.Username = m[datamodel.FieldQuestionUsername].(string)
		question.CreatedDate = utils.UnixTimeToTime(m[datamodel.FieldQuestionCreatedDate].(primitive.DateTime))

		if m[datamodel.FieldQuestionAnswer] != nil {
			ansArr := m[datamodel.FieldQuestionAnswer].(primitive.A)
			for _, a := range ansArr {
				ansDoc := a.(primitive.D)
				ansMap := ansDoc.Map()

				var ans datamodel.Answer
				ans.ID = ansMap[datamodel.FieldAnswerID].(primitive.ObjectID).Hex()
				if ansMap[datamodel.FieldAnswerIsGood] != nil {
					ans.IsGood = ansMap[datamodel.FieldAnswerIsGood].(bool)
				} else {
					ans.IsGood = false
				}

				question.Answers = append(question.Answers, ans)
			}
		} else {
			question.Answers = nil
		}

		//check if user already voted or not
		var voted = false
		if usr.Vote != nil {
			qObjId := m[datamodel.FieldQuestionID].(primitive.ObjectID)
			voted = isVoted(qObjId, usr.Vote)
		}

		//count number of answers
		numberOfAnswers := 0
		if question.Answers != nil {
			numberOfAnswers = len(question.Answers)
		}

		//check if the question is solved or not
		isSolved := false
		for _, elmt := range question.Answers {
			if elmt.IsGood {
				isSolved = true
				break
			}
		}

		q := QuestionsView{
			Question:        question,
			IsVoted:         voted,
			NumberOfAnswers: numberOfAnswers,
			IsSolved:        isSolved,
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
	var add int32
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
	username = utils.GetUsernameFromSession(s, r)

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

func isVoted(id primitive.ObjectID, voteArr primitive.A) bool {
	for _, elmt := range voteArr {
		if elmt == id {
			return true
		}
	}
	return false
}

func createPagination(r *http.Request) (int, []int, error) {
	var err error
	var curInt int
	//get page for pagination
	splitUrl := strings.Split(r.URL.Path, "/")
	if len(splitUrl) == 4 { // the url like /home/page/3, the first element is empty string : [,home,page,3]
		if splitUrl[2] == "page" {
			var cur int64
			cur, err = strconv.ParseInt(splitUrl[3], 10, 64)
			if err != nil {
				return 0, nil, err
			}
			curInt = int(cur)
		}
	}

	count := int64(0)
	count, err = db.Collection(datamodel.CollQuestion).CountDocuments(ctx, bson.D{}, nil)
	if err != nil {
		return 0, nil, err
	}

	pagesCount := int(math.Ceil(float64(count) / 5))
	pageIdx := make([]int, pagesCount)
	for i := 0; i < int(pagesCount); i++ {
		pageIdx[i] = i + 1
	}

	return curInt, pageIdx, nil
}
