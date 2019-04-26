package main

import (
	"context"

	"github.com/HotCodeGroup/warscript-utils/models"
	"github.com/pkg/errors"
)

type AuthManager struct{}

func (m *AuthManager) GetUserByID(ctx context.Context, userID *models.UserID) (*models.InfoUser, error) {
	usr, err := getInfoUserByIDImpl(userID.ID)
	if err != nil {
		return nil, errors.Wrap(err, "can not get user by id")
	}

	return &models.InfoUser{
		ID:        usr.ID,
		Username:  usr.Username,
		PhotoUUID: usr.PhotoUUID,
		Active:    usr.Active,
	}, nil
}

func (m *AuthManager) GetSessionInfo(ctx context.Context, token *models.SessionToken) (*models.SessionPayload, error) {
	payload, err := getSessionImpl(token.Token)
	if err != nil {
		return nil, errors.Wrap(err, "can not get session by token")
	}

	return &models.SessionPayload{
		ID: payload.ID,
	}, nil
}
