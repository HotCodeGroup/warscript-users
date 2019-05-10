package main

import (
	"context"
	"database/sql"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/HotCodeGroup/warscript-utils/utils"
	"github.com/pkg/errors"

	"github.com/HotCodeGroup/warscript-utils/logging"
	"github.com/HotCodeGroup/warscript-utils/models"
)

func init() {
	// выключаем логгер
	logger, _ = logging.NewLogger(ioutil.Discard, "")
}

func TestGetUserByID(t *testing.T) {
	m := &AuthManager{}

	Users = &UsersTest{
		ids: 1,
		users: map[int64]UserModel{
			1: {
				ID:       1,
				Username: "kek",
				Active:   true,
			},
		},
	}

	cases := []struct {
		id            int64
		expected      *models.InfoUser
		expectedError error
	}{
		{
			id: 1,
			expected: &models.InfoUser{
				ID:       1,
				Username: "kek",
				Active:   true,
			},
		},
		{
			id:            2,
			expectedError: utils.ErrNotExists,
		},
	}

	for i, c := range cases {
		req := &models.UserID{ID: c.id}
		resp, err := m.GetUserByID(context.Background(), req)
		if errors.Cause(err) != c.expectedError {
			t.Errorf("[%d] GetUserByIDTest got unexpected error: %v, expected: %v", i, err, c.expectedError)
		}
		if !reflect.DeepEqual(resp, c.expected) {
			t.Errorf("[%d] GetUserByIDTest returns: %v, wanted: %v", i, resp, c.expected)
		}
	}
}

func TestGetUserByUsername(t *testing.T) {
	m := &AuthManager{}

	Users = &UsersTest{
		ids: 1,
		users: map[int64]UserModel{
			1: {
				ID:        1,
				PhotoUUID: sql.NullString{String: "01010101-0101-0101-0101-010101010101", Valid: true},
				Username:  "kek",
				Active:    true,
			},
		},
	}

	cases := []struct {
		username      string
		expected      *models.InfoUser
		expectedError error
	}{
		{
			username: "kek",
			expected: &models.InfoUser{
				ID:        1,
				Username:  "kek",
				Active:    true,
				PhotoUUID: "01010101-0101-0101-0101-010101010101",
			},
		},
		{
			username:      "lol",
			expectedError: utils.ErrNotExists,
		},
	}

	for i, c := range cases {
		req := &models.Username{Username: c.username}
		resp, err := m.GetUserByUsername(context.Background(), req)
		if errors.Cause(err) != c.expectedError {
			t.Errorf("[%d] GetUserByIDTest got unexpected error: %v, expected: %v", i, err, c.expectedError)
		}
		if !reflect.DeepEqual(resp, c.expected) {
			t.Errorf("[%d] GetUserByIDTest returns: %v, wanted: %v", i, resp, c.expected)
		}
	}
}

func TestGetSessionInfo(t *testing.T) {
	m := &AuthManager{}

	Sessions = &SessionsTest{
		sessions: map[string][]byte{
			"1234": []byte(`{"id":1}`),
		},
	}

	cases := []struct {
		token         string
		expected      *models.SessionPayload
		expectedError error
	}{
		{
			token: "1234",
			expected: &models.SessionPayload{
				ID: 1,
			},
		},
		{
			token:         "1235",
			expectedError: utils.ErrNotExists,
		},
	}

	for i, c := range cases {
		req := &models.SessionToken{Token: c.token}
		resp, err := m.GetSessionInfo(context.Background(), req)
		if errors.Cause(err) != c.expectedError {
			t.Errorf("[%d] GetUserByIDTest got unexpected error: %v, expected: %v", i, err, c.expectedError)
		}
		if !reflect.DeepEqual(resp, c.expected) {
			t.Errorf("[%d] GetUserByIDTest returns: %v, wanted: %v", i, resp, c.expected)
		}
	}
}

func TestGetUsersByIDs(t *testing.T) {
	m := &AuthManager{}

	Users = &UsersTest{
		ids: 2,
		users: map[int64]UserModel{
			1: {
				ID:       1,
				Username: "kek",
				Active:   true,
			},
			2: {
				ID:        2,
				Username:  "kek1",
				PhotoUUID: sql.NullString{String: "01010101-0101-0101-0101-010101010101", Valid: true},
				Active:    true,
			},
		},
	}

	Users.(*UsersTest).SetNextFail(utils.ErrInternal)

	cases := []struct {
		ids           *models.UserIDs
		expected      *models.InfoUsers
		expectedError error
	}{
		{
			ids: &models.UserIDs{
				IDs: []*models.UserID{
					{ID: 1},
					{ID: 2},
					{ID: 3},
				},
			},
			expectedError: utils.ErrInternal,
		},
		{
			ids: &models.UserIDs{
				IDs: []*models.UserID{
					{ID: 1},
					{ID: 2},
					{ID: 3},
				},
			},
			expected: &models.InfoUsers{
				Users: []*models.InfoUser{
					{
						ID:       1,
						Username: "kek",
						Active:   true,
					},
					{
						ID:        2,
						Username:  "kek1",
						Active:    true,
						PhotoUUID: "01010101-0101-0101-0101-010101010101",
					},
				},
			},
		},
	}

	for i, c := range cases {
		resp, err := m.GetUsersByIDs(context.Background(), c.ids)
		if errors.Cause(err) != c.expectedError {
			t.Errorf("[%d] GetUserByIDTest got unexpected error: %v, expected: %v", i, err, c.expectedError)
		}
		if !reflect.DeepEqual(resp, c.expected) {
			t.Errorf("[%d] GetUserByIDTest returns: %v, wanted: %v", i, resp, c.expected)
		}
	}
}
