package discussion

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"text/template"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	datamodel "../../data_model"
	utils "../../utils"
)

var discussionTemplate = template.Must(template.ParseFiles(utils.WebsiteDirectory()+"/discussion/discussion.html",
	utils.WebsiteDirectory()+"/layout/main.html"))
var db *mongo.Database
var ctx context.Context

type DiscussionView struct {
	Question datamodel.Question
	IsVoted  bool
}

func DiscussionHandler(w http.ResponseWriter, r *http.Request) {
	err := initDiscussion()

	var qId primitive.ObjectID
	qId, err = getQuestionIdFromUrl(r)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "discussion : an error occured when retrieving question from url")
		return
	}

	if r.Method == "POST" {
		err = answer(r, qId)
		if err != nil {
			utils.InternalServerErrorHandler(w, r, err, "discussion : an error occured when answer")
			return
		}
	}

	var dView DiscussionView
	dView.Question, err = getQuestion(qId)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "discussion : an error occured when retrieving question")
		return
	}

	var voted = false
	voted, err = vote(r)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "discussion : an error occured when voting")
		return
	}

	if !voted {
		var usname string
		usname, err = utils.GetUsernameFromSession(r)
		if err != nil {
			utils.InternalServerErrorHandler(w, r, err, "discussion : an error occured when get username from session")
			return
		}

		var usr datamodel.User
		usr, err = getUser(usname)
		if err != nil {
			utils.InternalServerErrorHandler(w, r, err, "discussion : an error occured when retrieving user")
			return
		}

		if usr.Vote != nil {
			dView.IsVoted = isVoted(qId, usr.Vote)
		} else {
			dView.IsVoted = false
		}
	} else {
		dView.IsVoted = voted
	}

	err = discussionTemplate.ExecuteTemplate(w, "main.html", dView)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "discussion : an error occured when executing template")
		return
	}

}

func answer(r *http.Request, qId primitive.ObjectID) error {
	ans := r.FormValue("answer")
	usname, err := utils.GetUsernameFromSession(r)
	if err != nil {
		return err
	}

	var q datamodel.Question
	q, err = getQuestion(qId)
	if err != nil {
		return err
	}

	ansDoc := bson.D{
		{datamodel.FieldQuestionUsername, usname},
		{datamodel.FieldQuestionAnswer, ans},
		{datamodel.FieldQuestionCreatedDate, primitive.DateTime(timeToMillis(time.Now()))}}
	if q.Answer != nil {
		q.Answer = append(q.Answer, ansDoc)
	} else {
		q.Answer = bson.A{ansDoc}
	}

	qIdDoc := bson.D{{datamodel.FieldQuestionID, qId}}
	qUpdateDoc := bson.D{{datamodel.FieldQuestionAnswer, q.Answer}}
	_, err = db.Collection(datamodel.CollQuestion).UpdateOne(ctx, qIdDoc, bson.D{{"$set", qUpdateDoc}})
	if err != nil {
		return err
	}

	return nil
}

func timeToMillis(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func initDiscussion() error {
	var err error

	ctx := context.TODO()
	db, err = utils.ConnectDb(ctx)
	return err
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

func getQuestion(id primitive.ObjectID) (datamodel.Question, error) {
	qDoc := bson.D{}
	idDoc := bson.D{{datamodel.FieldQuestionID, id}}
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
		q.Answer = qMap[datamodel.FieldQuestionAnswer].(primitive.A)
	} else {
		q.Answer = nil
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

func isVoted(id primitive.ObjectID, voteArr primitive.A) bool {
	for _, elmt := range voteArr {
		if elmt == id {
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
	username, err = utils.GetUsernameFromSession(r)
	if err != nil {
		return false, err
	}

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
		return false, err
	}

	return true, nil
}
