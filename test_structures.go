package main

import (
	"github.com/HotCodeGroup/warscript-utils/testutils"
	"github.com/HotCodeGroup/warscript-utils/utils"
	"github.com/pkg/errors"
)

type usersTest struct {
	ids   int64
	users map[int64]UserModel

	testutils.Failer
}

func (u *usersTest) nextID() int64 {
	u.ids++
	return u.ids - 1
}

// Create создаёт запись в базе с новыми полями
func (u *usersTest) Create(m *UserModel) error {
	if err := u.NextFail(); err != nil {
		return err
	}

	m.Active = true
	m.ID = u.nextID()
	u.users[m.ID] = *m

	return nil
}

// Save сохраняет юзера в базу
func (u *usersTest) Save(m *UserModel) error {
	if err := u.NextFail(); err != nil {
		return err
	}

	u.users[m.ID] = *m
	return nil
}

// CheckPassword проверяет пароль у юзера и сохранённый в модели
func (u *usersTest) CheckPassword(m *UserModel, password string) bool {
	return *m.Password == password
}

// GetUserByID получает юзера по id
func (u *usersTest) GetUserBySecret(s string) (*UserModel, error) {
	if err := u.NextFail(); err != nil {
		return nil, err
	}

	return nil, nil
}

// GetUserByID получает юзера по id
func (u *usersTest) GetUserByID(id int64) (*UserModel, error) {
	if err := u.NextFail(); err != nil {
		return nil, err
	}

	m, ok := u.users[id]
	if !ok {
		return nil, utils.ErrNotExists
	}

	return &m, nil
}

// GetUserByUsername получает юзера по имени
func (u *usersTest) GetUserByUsername(username string) (*UserModel, error) {
	if err := u.NextFail(); err != nil {
		return nil, err
	}

	var m UserModel
	var ok bool

	for _, user := range u.users {
		if user.Username == username {
			ok = true
			m = user
		}
	}

	if !ok {
		return nil, utils.ErrNotExists
	}

	return &m, nil
}

// GetUsersByIDs получает юзеров по массиву айдишников
func (u *usersTest) GetUsersByIDs(ids []int64) ([]*UserModel, error) {
	if err := u.NextFail(); err != nil {
		return nil, err
	}

	users := make([]*UserModel, 0)
	for _, id := range ids {
		for uid := range u.users {
			if id == uid {
				usr := u.users[uid]
				users = append(users, &usr)
				break
			}
		}
	}

	return users, nil
}

type sessionsTest struct {
	sessions map[string][]byte

	testutils.Failer
}

// Set валидирует и сохраняет сессию в хранилище по сгенерированному токену
// Токен сохраняется в s.Token
func (ss *sessionsTest) Set(s *Session) error {
	if err := ss.NextFail(); err != nil {
		return err
	}

	ss.sessions[s.Token] = s.Payload
	return nil
}

// Delete удаляет сессию с токен s.Token из хранилища
func (ss *sessionsTest) Delete(s *Session) error {
	if err := ss.NextFail(); err != nil {
		return err
	}
	delete(ss.sessions, s.Token)

	return nil
}

// GetSession получает сессию из хранилища по токену
func (ss *sessionsTest) GetSession(token string) (*Session, error) {
	if err := ss.NextFail(); err != nil {
		return nil, err
	}
	data, ok := ss.sessions[token]
	if !ok {
		return nil, errors.Wrap(utils.ErrNotExists, "redis get error")
	}

	return &Session{
		Token:   token,
		Payload: data,
	}, nil
}
