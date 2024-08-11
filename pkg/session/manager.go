package session

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"

	"reddit/pkg/user"
)

type SessionManagerDataBase interface {
	CreateSessionDB(newSession *Session, token string) error
	GetSessionDB(inToken string) (*Session, error)
}

type SessionManager struct {
	mu           *sync.RWMutex
	SessionManDB SessionManagerDataBase
	secret       []byte
}

func NewSessionManager(sessManDB SessionManagerDB) *SessionManager {
	mySecret := os.Getenv("SECRET")
	return &SessionManager{
		mu:           &sync.RWMutex{},
		SessionManDB: &sessManDB,
		secret:       []byte(mySecret),
	}
}

func (sm *SessionManager) newToken(user *user.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": user,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Hour * 24 * 7).Unix(),
	})
	tokenString, err := token.SignedString(sm.secret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (sm *SessionManager) CreateNewSession(u *user.User) (string, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	token, err := sm.newToken(u)
	if err != nil {
		return "", err
	}
	newSession := newSession(u, token)
	err = sm.SessionManDB.CreateSessionDB(newSession, token)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (sm *SessionManager) GetSession(inToken string) (*Session, error) {
	hashSecretGetter := func(token *jwt.Token) (interface{}, error) {
		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok || method.Alg() != "HS256" {
			fmt.Println("bad sign in")
			return nil, fmt.Errorf("bad sign method")
		}
		return sm.secret, nil
	}
	token, err := jwt.Parse(inToken, hashSecretGetter)
	if err != nil || !token.Valid {
		fmt.Println("bad secret")

		return nil, ErrNoAuth
	}
	sm.mu.RLock()
	sess, err := sm.SessionManDB.GetSessionDB(inToken)
	sm.mu.RUnlock()
	if err != nil {
		return nil, err
	}
	return sess, nil
}
