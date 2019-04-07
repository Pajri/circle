package discussion

import (
	"context"
	"net/http"
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

func DiscussionHandler(w http.ResponseWriter, r *http.Request) {
	err := initDiscussion()
	err = discussionTemplate.ExecuteTemplate(w, "main.html", nil)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "login : an error occured when executing template")
		return
	}
}

func initDiscussion() error {
	var err error

	ctx := context.TODO()
	db, err = utils.ConnectDb(ctx)
	return err
}

func getDiscussion(id primitive.ObjectID) (datamodel.Question, error) {
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
