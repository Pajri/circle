package datamodel

import "time"

var FieldAnswerID = "_id"
var FieldAnswerAnswer = "answer"
var FieldAnswerUsername = "username"
var FieldAnswerIsGood = "isGood"
var FieldAnswerCreatedDate = "createdDate"

type Answer struct {
	ID          string
	Answer      string
	Username    string
	IsGood      bool
	CreatedDate time.Time
	ImageUrl    string
}
