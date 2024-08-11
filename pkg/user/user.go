package user

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	password string
}

type UserRepo interface {
	Login(username, password string) (*User, error)
	Register(username, password string) (*User, error)
}

func newUser(id, uName, pass string) *User {
	return &User{
		ID:       id,
		Username: uName,
		password: pass,
	}
}
