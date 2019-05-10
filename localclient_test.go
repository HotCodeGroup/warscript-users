package main

import (
	"context"
	"reflect"
	"testing"

	"github.com/HotCodeGroup/warscript-utils/models"
	"github.com/HotCodeGroup/warscript-utils/utils"
)

func TestLocalAuthClientOK(t *testing.T) {
	c := &LocalAuthClient{}

	Sessions = &sessionsTest{
		sessions: map[string][]byte{
			"1234": []byte(`{"id":1}`),
		},
	}

	payload, err := c.GetSessionInfo(context.Background(), &models.SessionToken{
		Token: "1234",
	})
	if err != nil {
		t.Errorf("TestLocalAuthClientOK got unexpected error: %v", err)
	}

	expected := &models.SessionPayload{
		ID: 1,
	}
	if !reflect.DeepEqual(payload, expected) {
		t.Errorf("TestLocalAuthClientOK got unexpected result: %v, expected: %v", payload, expected)
	}
}

func TestLocalAuthClientErr(t *testing.T) {
	c := &LocalAuthClient{}

	Sessions = &sessionsTest{
		sessions: map[string][]byte{
			"1234": []byte(`{"id":1}`),
		},
	}

	Sessions.(*sessionsTest).SetNextFail(utils.ErrInternal)

	_, err := c.GetSessionInfo(context.Background(), &models.SessionToken{
		Token: "1234",
	})
	if err != utils.ErrInternal {
		t.Errorf("TestLocalAuthClientOK got unexpected error: %v, expected: %v", err, utils.ErrInternal)
	}
}

func TestLocalAuthClientMocks(t *testing.T) {
	c := &LocalAuthClient{}

	if _, err := c.GetUserByID(context.Background(), &models.UserID{ID: 1}); err != nil {
		t.Errorf("TestLocalAuthClientMocks got unexpected error: %v", err)
	}

	if _, err := c.GetUserByUsername(context.Background(), &models.Username{Username: "kek"}); err != nil {
		t.Errorf("TestLocalAuthClientMocks got unexpected error: %v", err)
	}

	if _, err := c.GetUsersByIDs(context.Background(), &models.UserIDs{
		IDs: []*models.UserID{
			{ID: 1},
			{ID: 2},
			{ID: 3},
		},
	}); err != nil {
		t.Errorf("TestLocalAuthClientMocks got unexpected error: %v", err)
	}
}
