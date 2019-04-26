package main

import (
	"encoding/json"
	"time"

	"github.com/HotCodeGroup/warscript-utils/utils"
	"github.com/pkg/errors"
)

func createSessionImpl(form *FormUser) (*Session, error) {
	if err := form.Validate(); err != nil {
		return nil, err
	}

	user, err := Users.GetUserByUsername(form.Username)
	if err != nil {
		return nil, &utils.ValidationError{
			"username": utils.ErrNotExists.Error(),
		}
	}

	if !Users.CheckPassword(user, form.Password) {
		return nil, &utils.ValidationError{
			"password": utils.ErrInvalid.Error(),
		}
	}

	data, err := json.Marshal(&SessionPayload{
		ID: user.ID.Int,
	})
	if err != nil {
		return nil, errors.Wrap(err, "info marshal error")
	}

	session := &Session{
		Payload:      data,
		ExpiresAfter: time.Hour * 24 * 30,
	}
	err = Sessions.Set(session)
	if err != nil {
		return nil, errors.Wrap(err, "set session error")
	}

	return session, nil
}

func getSessionImpl(token string) (*SessionPayload, error) {
	session, err := Sessions.GetSession(token)
	if err != nil {
		return nil, err
	}
	payload := &SessionPayload{}
	err = json.Unmarshal(session.Payload, payload)
	if err != nil {
		return nil, errors.Wrap(err, "payload unmarshal error")
	}

	return payload, nil
}
