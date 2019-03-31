package datamodel

var CollQuestion = "question"

type Question struct {
	ID          string `json:_id`
	Title       string `json:title`
	Description string `json:description`
	Vote        int    `json:vote`
	IsSolved    bool   `json:isSolved`
	Username    string `json:username`
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
