package session

import (
	"encoding/json"

	"github.com/gomodule/redigo/redis"
)

const expireTime = 24 * 60 * 60

type SessionManagerRedis struct {
	RedisConn redis.Conn
}

func (sm *SessionManagerRedis) CreateSession(newSession *Session, token string) error {
	sessionJSON, err := json.Marshal(newSession)
	if err != nil {
		return err
	}
	result, err := redis.String(sm.RedisConn.Do("SET", token, sessionJSON, "EX", expireTime))
	if err != nil || result != "OK" {
		return err
	}
	return nil

}

func (sm *SessionManagerRedis) GetSession(inToken string) (*Session, error) {
	sess := &Session{}
	sessFromRedis, err := redis.String(sm.RedisConn.Do("GET", inToken))
	if err != nil {
		return nil, ErrNoAuth
	}
	err = json.Unmarshal([]byte(sessFromRedis), sess)
	if err != nil {
		return nil, ErrNoAuth
	}
	return sess, nil
}
