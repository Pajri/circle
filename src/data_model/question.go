package datamodel

import (
	"time"
)

var CollQuestion = "question"
var FieldQuestionID = "_id"
var FieldQuestionCreatedDate = "createdDate"
var FieldQuestionVote = "vote"

type Question struct {
	ID          string    `json:_id`
	Title       string    `json:title`
	Description string    `json:description`
	Vote        int       `json:vote`
	IsSolved    bool      `json:isSolved`
	Username    string    `json:username`
	CreatedDate time.Time `json:createdDate`
}

func (question Question) IDColl() string {
	return "_id"
}

func (question Question) TitleColl() string {
	return "title"
}

func (question Question) DescriptionColl() string {
	return "description"
}

func (question Question) VoteColl() string {
	return "vote"
}

func (question Question) IsSolvedColl() string {
	return "isSolved"
}

func (question Question) UsernameColl() string {
	return "username"
}

func (question Question) CreatedDateColl() string {
	return "createdDate"
}
