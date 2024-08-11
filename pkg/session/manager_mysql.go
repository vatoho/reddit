package session

import (
	"database/sql"
	"errors"
	"fmt"

	"reddit/pkg/user"
)

type SessionManagerMysql struct {
	DB *sql.DB
}

func (sm *SessionManagerMysql) CreateSession(newSession *Session, token string) error {
	_, err := sm.DB.Exec("INSERT INTO sessions(`token`, `user_id`) VALUES (?, ?)",
		token, newSession.User.ID)
	return err
}

func (sm *SessionManagerMysql) GetSession(inToken string) (*Session, error) {
	currentSession := &Session{}
	userForSession := &user.User{}
	err := sm.DB.QueryRow("SELECT token, id, username FROM sessions JOIN users ON sessions.user_id = users.id WHERE sessions.token = ?", inToken).
		Scan(&currentSession.ID, &userForSession.ID, &userForSession.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("Запись не найдена")
		} else {
			fmt.Printf("Произошла ошибка: %v\n", err)
		}
		fmt.Println("not found in db")
		return nil, ErrNoAuth
	}
	currentSession.User = userForSession
	return currentSession, nil
}
