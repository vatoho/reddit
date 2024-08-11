package session

import "fmt"

type SessionManagerDB struct {
	SessionManagerMS  SessionManagerMysql
	SessionManagerRDS SessionManagerRedis
}

func (sm *SessionManagerDB) CreateSessionDB(newSession *Session, token string) error {

	err := sm.SessionManagerRDS.CreateSession(newSession, token)
	if err != nil {
		fmt.Println("can not create session in redis")
	}
	err = sm.SessionManagerMS.CreateSession(newSession, token)
	return err
}

func (sm *SessionManagerDB) GetSessionDB(inToken string) (*Session, error) {
	sess, err := sm.SessionManagerRDS.GetSession(inToken)
	if err == nil && sess != nil {
		fmt.Println("session from redis")
		return sess, nil
	} else {
		fmt.Println("no session in redis")
	}
	sess, err = sm.SessionManagerMS.GetSession(inToken)
	if err != nil {
		return nil, err
	}
	return sess, nil
}
