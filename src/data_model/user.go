package datamodel

var CollUser = "user"
var FieldUserUsername = "username"

type User struct {
	ID       string   `json:_id`
	Username string   `json:username`
	Email    string   `json:email`
	Password string   `json:password`
	Votes    []string `json:votes`
}

func (user User) CollName() string {
	return "user"
}

func (user User) UsernameColl() string {
	return "username"
}

func (user User) EmailColl() string {
	return "email"
}

func (user User) PasswordColl() string {
	return "password"
}
