package main

import (
	"context"

	"github.com/HotCodeGroup/warscript-utils/models"
	"google.golang.org/grpc"
)

type LocalAuthClient struct{}

func (c *LocalAuthClient) GetUserByID(ctx context.Context,
	in *models.UserID, opts ...grpc.CallOption) (*models.InfoUser, error) {
	return nil, nil
}

func (c *LocalAuthClient) GetUserByUsername(ctx context.Context,
	in *models.Username, opts ...grpc.CallOption) (*models.InfoUser, error) {
	return nil, nil
}

func (c *LocalAuthClient) GetUserBySecret(ctx context.Context,
	in *models.VkSecret, opts ...grpc.CallOption) (*models.InfoUser, error) {
	return nil, nil
}

func (c *LocalAuthClient) GetSessionInfo(ctx context.Context,
	in *models.SessionToken, opts ...grpc.CallOption) (*models.SessionPayload, error) {
	payload, err := getSessionImpl(in.Token)
	if err != nil {
		return nil, err
	}

	return &models.SessionPayload{
		ID: payload.ID,
	}, nil
}

func (c *LocalAuthClient) GetUsersByIDs(ctx context.Context,
	in *models.UserIDs, opts ...grpc.CallOption) (*models.InfoUsers, error) {
	return nil, nil
}
