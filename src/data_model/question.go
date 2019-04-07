package datamodel

import (
	"time"
)

var CollQuestion = "question"
var FieldQuestionID = "_id"
var FieldQuestionCreatedDate = "createdDate"
var FieldQuestionVote = "vote"

type Question struct {
	ID          string
	Title       string
	Description string
	Vote        int
	IsSolved    bool
	Username    string
	CreatedDate time.Time
}

func (q Question) IDColl() string {
	return "_id"
}

func (q Question) TitleColl() string {
	return "title"
}

func (q Question) DescriptionColl() string {
	return "description"
}

func (q Question) VoteColl() string {
	return "vote"
}

func (q Question) IsSolvedColl() string {
	return "isSolved"
}

func (q Question) UsernameColl() string {
	return "username"
}

func (q Question) FormattedCreatedDate() string {
	return q.CreatedDate.Format("1 Jan 2001")
}
