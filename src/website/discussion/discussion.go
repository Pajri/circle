package discussion

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"

	"github.com/gorilla/sessions"

	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	datamodel "../../data_model"
	utils "../../utils"
)

var funcMap = template.FuncMap{"formatCreatedDate": utils.FormatCreatedDate}
var discussionTemplate = template.Must(template.New("discussion.html").Funcs(funcMap).ParseFiles(utils.WebsiteDirectory()+"/discussion/discussion.html",
	utils.WebsiteDirectory()+"/layout/main.html"))
var db *mongo.Database
var ctx context.Context
var questionId primitive.ObjectID
var s *sessions.Session

type DiscussionView struct {
	Question                 datamodel.Question
	IsVoted                  bool
	NumOfAnswer              int
	IsQuestionByLoggedInUser bool
}

func DiscussionHandler(w http.ResponseWriter, r *http.Request) {
	err := initDiscussion(r)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "discussion : an error occured when initializing discussion")
		return
	}

	//check is logged in
	isAuth := utils.IsLoggedInSession(s)
	if !isAuth {
		utils.ForbiddenHandler(w, r)
		return
	}

	questionId, err = getQuestionIdFromUrl(r)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "discussion : an error occured when getting question from url")
		return
	}

	if r.Method == "POST" {
		err = answer(r)
		if err != nil {
			utils.InternalServerErrorHandler(w, r, err, "discussion : an error occured when answer")
			return
		}

		err = setGoodAnswer(r)
		if err != nil {
			utils.InternalServerErrorHandler(w, r, err, "discussion : an error occured when set good answer")
			return
		}
	}

	var dView DiscussionView
	dView.Question, err = getQuestion()
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "discussion : an error occured when retrieving question")
		return
	}

	dView.NumOfAnswer = 0
	if dView.Question.Answers != nil {
		dView.NumOfAnswer = len(dView.Question.Answers)
	}
	var voted = false
	voted, err = vote(r)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "discussion : an error occured when voting")
		return
	}

	if !voted {
		var usname string
		usname = utils.GetUsernameFromSession(s, r)

		var usr datamodel.User
		usr, err = getUser(usname)
		if err != nil {
			utils.InternalServerErrorHandler(w, r, err, "discussion : an error occured when retrieving user")
			return
		}

		if usr.Vote != nil {
			dView.IsVoted = isVoted(usr.Vote)
		} else {
			dView.IsVoted = false
		}
	} else {
		dView.IsVoted = voted
	}

	//should show solve button or not
	var loggedInUser string
	loggedInUser = utils.GetUsernameFromSession(s, r)

	dView.IsQuestionByLoggedInUser = false
	if loggedInUser == dView.Question.Username {
		dView.IsQuestionByLoggedInUser = true
	}

	err = discussionTemplate.ExecuteTemplate(w, "main.html", dView)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "discussion : an error occured when executing template")
		return
	}

	db.Client().Disconnect(ctx)
}

func initDiscussion(r *http.Request) error {
	var err error

	ctx := context.TODO()
	db, err = utils.ConnectDb(ctx)
	if err != nil {
		return err
	}

	s, err = utils.GetSession(r, utils.SESSION_AUTH) //init session
	if err != nil {
		return err
	}
	return nil
}

func answer(r *http.Request) error {
	var err error
	ans := r.FormValue("answer")
	if ans == "" { //the post request is not for answering question
		return nil
	}

	usname := utils.GetUsernameFromSession(s, r)

	var q datamodel.Question
	q, err = getQuestion()
	if err != nil {
		return err
	}

	newAnsDoc := bson.D{
		{datamodel.FieldAnswerID, primitive.NewObjectID()},
		{datamodel.FieldAnswerAnswer, ans},
		{datamodel.FieldAnswerUsername, usname},
		{datamodel.FieldQuestionCreatedDate, primitive.DateTime(utils.TimeToMillis(time.Now()))}}

	answersArr := bson.A{}
	if q.Answers != nil {
		//add answers
		for _, ans := range q.Answers {
			ansObjID, _ := primitive.ObjectIDFromHex(ans.ID)
			ansDoc := bson.D{
				{datamodel.FieldAnswerID, ansObjID},
				{datamodel.FieldAnswerAnswer, ans.Answer},
				{datamodel.FieldAnswerUsername, ans.Username},
				{datamodel.FieldAnswerCreatedDate, primitive.DateTime(utils.TimeToMillis(ans.CreatedDate))},
			}
			answersArr = append(answersArr, ansDoc) //list stored answer
		}
		answersArr = append(answersArr, newAnsDoc) //add new answer
	} else {
		answersArr = append(answersArr, newAnsDoc)
	}

	qIdDoc := bson.D{{datamodel.FieldQuestionID, questionId}}
	qUpdateDoc := bson.D{{datamodel.FieldQuestionAnswer, answersArr}}
	_, err = db.Collection(datamodel.CollQuestion).UpdateOne(ctx, qIdDoc, bson.D{{"$set", qUpdateDoc}})
	if err != nil {
		return err
	}

	return nil
}

func getQuestionIdFromUrl(r *http.Request) (primitive.ObjectID, error) {
	var objId primitive.ObjectID
	var err error

	urlSplit := strings.Split(r.URL.Path, "/")
	if len(urlSplit) != 3 {
		return objId, fmt.Errorf("url does not contain question id")
	}

	id := urlSplit[2]
	objId, err = primitive.ObjectIDFromHex(id)
	if err != nil {
		return objId, err
	}

	return objId, nil
}

func getQuestion() (datamodel.Question, error) {
	qDoc := bson.D{}
	idDoc := bson.D{{datamodel.FieldQuestionID, questionId}}
	err := db.Collection(datamodel.CollQuestion).FindOne(ctx, idDoc).Decode(&qDoc)
	if err != nil {
		return datamodel.Question{}, err
	}
	qMap := qDoc.Map()

	var q datamodel.Question
	q.ID = qMap[datamodel.FieldQuestionID].(primitive.ObjectID).Hex()
	q.Title = qMap[datamodel.FieldQuestionTitle].(string)
	q.Description = qMap[datamodel.FieldQuestionDescription].(string)
	q.Vote = qMap[datamodel.FieldQuestionVote].(int32)
	q.IsSolved = qMap[datamodel.FieldQuestionIsSolved].(bool)
	q.Username = qMap[datamodel.FieldQuestionUsername].(string)
	q.CreatedDate = utils.UnixTimeToTime(qMap[datamodel.FieldQuestionCreatedDate].(primitive.DateTime))

	if qMap[datamodel.FieldQuestionAnswer] != nil {
		ansArr := qMap[datamodel.FieldQuestionAnswer].(primitive.A)
		for _, a := range ansArr {
			ansDoc := a.(primitive.D)
			ansMap := ansDoc.Map()

			var ans datamodel.Answer
			ans.ID = ansMap[datamodel.FieldAnswerID].(primitive.ObjectID).Hex()
			ans.Answer = ansMap[datamodel.FieldAnswerAnswer].(string)
			ans.Username = ansMap[datamodel.FieldAnswerUsername].(string)
			ans.CreatedDate = utils.UnixTimeToTime(ansMap[datamodel.FieldAnswerCreatedDate].(primitive.DateTime))

			if ansMap[datamodel.FieldAnswerIsGood] != nil {
				ans.IsGood = ansMap[datamodel.FieldAnswerIsGood].(bool)
			} else {
				ans.IsGood = false
			}

			q.Answers = append(q.Answers, ans)
		}
	} else {
		q.Answers = nil
	}

	return q, nil
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

func isVoted(voteArr primitive.A) bool {
	for _, elmt := range voteArr {
		if elmt == questionId {
			return true
		}
	}
	return false
}

func vote(r *http.Request) (bool, error) {
	var ok bool
	var val []string
	val, ok = r.URL.Query()["action"]
	if !ok {
		//action not passed, no voting
		return false, nil
	}
	action := val[0]
	var add int32
	if action == "upvote" {
		add = 11
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
		return false, err
	}

	updateDoc := bson.D{{datamodel.FieldQuestionVote, q.Vote + add}} //update vote data
	_, err = db.Collection(datamodel.CollQuestion).UpdateOne(ctx, idDoc, bson.D{{"$set", updateDoc}})
	if err != nil {
		return false, err
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
		return false, err
	}

	//create vote array
	m := usrDoc.Map()
	var voteArray bson.A
	if m[datamodel.FieldUserVote] != nil {
		voteArray = m[datamodel.FieldUserVote].(bson.A)
		if !isVoted(voteArray) {
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
		return false, err
	}

	return true, nil
}

func setGoodAnswer(r *http.Request) error {
	ansId := r.FormValue("good-answer")
	if ansId == "" {
		return nil
	}

	q, err := getQuestion()
	if err != nil {
		return err
	}

	actionTaken := 0
	for i := 0; i < len(q.Answers); i++ {
		if q.Answers[i].ID != ansId && q.Answers[i].IsGood {
			q.Answers[i].IsGood = false
			actionTaken++
		}

		if q.Answers[i].ID == ansId {
			q.Answers[i].IsGood = true
			actionTaken++
		}

		if actionTaken == 2 {
			break
		}
	}

	//create ans array
	ansArr := bson.A{}
	for _, elmt := range q.Answers {
		ansId, _ := primitive.ObjectIDFromHex(elmt.ID)
		ansDoc := bson.D{
			{datamodel.FieldAnswerID, ansId},
			{datamodel.FieldAnswerAnswer, elmt.Answer},
			{datamodel.FieldAnswerUsername, elmt.Username},
			{datamodel.FieldAnswerCreatedDate, primitive.DateTime(utils.TimeToMillis(elmt.CreatedDate))},
			{datamodel.FieldAnswerIsGood, elmt.IsGood},
		}
		ansArr = append(ansArr, ansDoc)
	}

	qIdDoc := bson.D{{datamodel.FieldQuestionID, questionId}}
	ansUpdateDoc := bson.D{{datamodel.FieldQuestionAnswer, ansArr}}
	_, err = db.Collection(datamodel.CollQuestion).UpdateOne(ctx, qIdDoc, bson.D{{"$set", ansUpdateDoc}})
	if err != nil {
		return err
	}

	return nil
}
