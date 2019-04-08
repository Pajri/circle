package datamodel

import (
	"time"
)

var CollQuestion = "question"
var FieldQuestionID = "_id"
var FieldQuestionTitle = "title"
var FieldQuestionDescription = "description"
var FieldQuestionVote = "vote"
var FieldQuestionIsSolved = "isSolved"
var FieldQuestionUsername = "username"
var FieldQuestionCreatedDate = "createdDate"

type Question struct {
	ID          string
	Title       string
	Description string
	Vote        int
	IsSolved    bool
	Username    string
	CreatedDate time.Time
}
