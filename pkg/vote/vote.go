package vote

type Vote struct {
	Value  int    `json:"vote"`
	UserID string `json:"user"`
}

func NewVote(value int, userID string) *Vote {
	return &Vote{
		Value:  value,
		UserID: userID,
	}
}
