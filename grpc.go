package main

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/HotCodeGroup/warscript-utils/models"
	"github.com/pkg/errors"
)

type AuthManager struct{}

func (m *AuthManager) GetUserByID(ctx context.Context, userID *models.UserID) (*models.InfoUser, error) {
	logger := logger.WithFields(logrus.Fields{
		"method":  "grpc_GetUserByID",
		"user_id": userID.ID,
	})

	usr, err := getInfoUserByIDImpl(userID.ID)
	if err != nil {
		logger.Errorf("can not get user by id: %s", err)
		return nil, errors.Wrap(err, "can not get user by id")
	}

	logger.Info("successfull")
	return &models.InfoUser{
		ID:        usr.ID,
		Username:  usr.Username,
		PhotoUUID: usr.PhotoUUID,
		Active:    usr.Active,
	}, nil
}

func (m *AuthManager) GetSessionInfo(ctx context.Context, token *models.SessionToken) (*models.SessionPayload, error) {
	logger := logger.WithFields(logrus.Fields{
		"method": "grpc_GetSessionInfo",
		"token":  token.Token,
	})

	payload, err := getSessionImpl(token.Token)
	if err != nil {
		logger.Errorf("can not get session by token: %s", err)
		return nil, errors.Wrap(err, "can not get session by token")
	}

	logger.Info("successfull")
	return &models.SessionPayload{
		ID: payload.ID,
	}, nil
}

func (m *AuthManager) GetUsersByIDs(context.Context, *models.UserIDs) (*models.InfoUsers, error) {
	return nil, nil
}
