package datamodel

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var CollQuestion = "question"
var FieldQuestionID = "_id"
var FieldQuestionTitle = "title"
var FieldQuestionDescription = "description"
var FieldQuestionVote = "vote"
var FieldQuestionIsSolved = "isSolved"
var FieldQuestionUsername = "username"
var FieldQuestionAnswer = "answer"
var FieldQuestionCreatedDate = "createdDate"

type Question struct {
	ID          string
	Title       string
	Description string
	Vote        int32
	IsSolved    bool
	Username    string
	Answer      primitive.A
	CreatedDate time.Time
}
