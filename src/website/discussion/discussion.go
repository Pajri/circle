package discussion

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"text/template"

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
}

func DiscussionHandler(w http.ResponseWriter, r *http.Request) {
	err := initDiscussion()

	var qId primitive.ObjectID
	qId, err = getQuestionIdFromUrl(r)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "discussion : an error occured when retrieving question from url")
		return
	}

	var dView DiscussionView
	dView.Question, err = getQuestion(qId)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "discussion : an error occured when retrieving question")
		return
	}

	err = discussionTemplate.ExecuteTemplate(w, "main.html", dView)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "discussion : an error occured when executing template")
		return
	}

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

	id := urlSplit[3]
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
	q.Vote = qMap[datamodel.FieldQuestionVote].(int)
	q.IsSolved = qMap[datamodel.FieldQuestionIsSolved].(bool)
	q.Username = qMap[datamodel.FieldQuestionUsername].(string)
	q.CreatedDate = utils.UnixTimeToTime(qMap[datamodel.FieldQuestionCreatedDate].(primitive.DateTime))

	return q, nil
}
