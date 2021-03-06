package main

import (
	"time"

	"github.com/HotCodeGroup/warscript-utils/utils"
	"github.com/go-redis/redis"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var rediCli *redis.Client

// SessionAccessObject DAO for Session model
type SessionAccessObject interface {
	Set(s *Session) error
	Delete(s *Session) error
	GetSession(token string) (*Session, error)
}

// SessionConn implementation of SessionAccessObject
type SessionConn struct{}

// Sessions interface variable for models methods
var Sessions SessionAccessObject

func init() {
	Sessions = &SessionConn{}
}

// Session модель для работы с сессиями
type Session struct {
	Token        string
	Payload      []byte
	ExpiresAfter time.Duration
}

// Set валидирует и сохраняет сессию в хранилище по сгенерированному токену
// Токен сохраняется в s.Token
func (ss *SessionConn) Set(s *Session) error {
	sessionToken := uuid.New()
	err := rediCli.Set(sessionToken.String(), s.Payload, s.ExpiresAfter).Err()
	if err != nil {
		return errors.Wrapf(utils.ErrInternal, "redis save error: %v", err)
	}

	s.Token = sessionToken.String()
	return nil
}

// Delete удаляет сессию с токен s.Token из хранилища
func (ss *SessionConn) Delete(s *Session) error {
	err := rediCli.Del(s.Token).Err()
	if err != nil {
		return errors.Wrapf(utils.ErrInternal, "redis delete error: %v", err)
	}

	return nil
}

// GetSession получает сессию из хранилища по токену
func (ss *SessionConn) GetSession(token string) (*Session, error) {
	data, err := rediCli.Get(token).Bytes()
	if err != nil {
		return nil, errors.Wrapf(utils.ErrInternal, "redis get error: %v", err)
	}

	return &Session{
		Token:   token,
		Payload: data,
	}, nil
}
