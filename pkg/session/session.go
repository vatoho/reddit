package session

import (
	"errors"

	"reddit/pkg/user"
)

type SessManager interface {
	CreateNewSession(u *user.User) (string, error)
	GetSession(inToken string) (*Session, error)
}

type Session struct {
	ID   string
	User *user.User
}

var (
	ErrNoAuth = errors.New("no session found")
)

func newSession(user *user.User, id string) *Session {
	return &Session{
		ID:   id,
		User: user,
	}
}
