package datamodel

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var CollUser = "user"
var FieldUserID = "_id"
var FieldUserUsername = "username"
var FieldUserEmail = "email"
var FieldPassword = "password"
var FieldUserVote = "vote"

type User struct {
	ID       string
	Username string
	Email    string
	Password string
	Vote     primitive.A
}
