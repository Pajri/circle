package datamodel

import "time"

var FieldAnswerID = "_id"
var FieldAnswerAnswer = "answer"
var FieldAnswerUsername = "username"
var FieldAnswerCreatedDate = "createdDate"

type Answer struct {
	ID          string
	Answer      string
	Username    string
	CreatedDate time.Time
}
