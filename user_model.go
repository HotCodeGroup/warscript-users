package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/HotCodeGroup/warscript-utils/postgresql"
	"github.com/HotCodeGroup/warscript-utils/utils"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"

	"database/sql"

	_ "github.com/lib/pq"
)

var pqConn *sql.DB

// UserAccessObject DAO for User model
type UserAccessObject interface {
	GetUserByID(id int64) (*UserModel, error)
	GetUserByUsername(username string) (*UserModel, error)
	GetUsersByIDs(ids []int64) ([]*UserModel, error)

	Create(u *UserModel) error
	Save(u *UserModel) error
	CheckPassword(u *UserModel, password string) bool
}

// AccessObject implementation of UserAccessObject
type AccessObject struct{}

// Users interface variable for models methods
var Users UserAccessObject

func init() {
	Users = &AccessObject{}
}

// UserModel model for users table
type UserModel struct {
	ID            int64
	Username      string
	PhotoUUID     sql.NullString
	Password      *string // строка для сохранения
	Active        bool
	PasswordCrypt []byte // внутренний хеш для проверки
}

// GetPhotoUUID возвращает photoUUID или пустую строку, если его нет в базе
func (u *UserModel) GetPhotoUUID() string {
	if u.PhotoUUID.Valid {
		return u.PhotoUUID.String
	}

	return ""
}

// Create создаёт запись в базе с новыми полями
func (us *AccessObject) Create(u *UserModel) error {
	var err error
	u.PasswordCrypt, err = bcrypt.GenerateFromPassword([]byte(*u.Password), bcrypt.MinCost)
	if err != nil {
		return errors.Wrapf(utils.ErrInternal, "password generate error: %s", err.Error())
	}

	tx, err := pqConn.Begin()
	if err != nil {
		return errors.Wrapf(utils.ErrInternal, "can not open user create transaction: %s", err.Error())
	}
	//nolint:errcheck
	defer tx.Rollback()

	_, err = us.getUserImpl(tx, "username", u.Username)
	if err != sql.ErrNoRows {
		if err == nil {
			return utils.ErrTaken
		}

		return errors.Wrapf(utils.ErrInternal, "check duplicate error: %s", err.Error())
	}

	_, err = tx.Exec(`INSERT INTO users (username, password) VALUES($1, $2);`, &u.Username, &u.PasswordCrypt)
	if err != nil {
		return errors.Wrapf(utils.ErrInternal, "user create error: %s", err.Error())
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrapf(utils.ErrInternal, "user create transaction commit error: %s", err.Error())
	}

	return nil
}

// Save сохраняет юзера в базу
func (us *AccessObject) Save(u *UserModel) error {
	var err error
	if u.Password != nil {
		u.PasswordCrypt, err = bcrypt.GenerateFromPassword([]byte(*u.Password), bcrypt.MinCost)
		if err != nil {
			return errors.Wrapf(utils.ErrInternal, "password generate error: %s", err.Error())
		}
	}

	tx, err := pqConn.Begin()
	if err != nil {
		return errors.Wrapf(utils.ErrInternal, "can not open user save transaction: %s", err.Error())
	}
	//nolint:errcheck
	defer tx.Rollback()

	du, err := us.getUserImpl(tx, "username", u.Username)
	if err == nil && u.ID != du.ID {
		return utils.ErrTaken
	} else if err != nil && err != sql.ErrNoRows {
		return errors.Wrapf(utils.ErrInternal, "check duplicate error: %s", err.Error())
	}

	_, err = tx.Exec(`UPDATE users SET (username, password, photo_uuid, active) = (
		COALESCE($1, username),
		COALESCE($2, password),
		$3,
		COALESCE($4, active)
		)
		WHERE id = $5;`,
		&u.Username, &u.PasswordCrypt, &u.PhotoUUID, &u.Active, &u.ID)
	if err != nil {
		return errors.Wrapf(utils.ErrInternal, "user save error: %s", err.Error())
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrapf(utils.ErrInternal, "user save transaction commit error: %s", err.Error())
	}

	return nil
}

// CheckPassword проверяет пароль у юзера и сохранённый в модели
func (us *AccessObject) CheckPassword(u *UserModel, password string) bool {
	err := bcrypt.CompareHashAndPassword(u.PasswordCrypt, []byte(password))
	return err == nil
}

// GetUserByID получает юзера по id
func (us *AccessObject) GetUserByID(id int64) (*UserModel, error) {
	u, err := us.getUserImpl(pqConn, "id", strconv.FormatInt(id, 10))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrNotExists
		}

		return nil, errors.Wrapf(utils.ErrInternal, "user get by id error: %s", err.Error())
	}

	return u, nil
}

// GetUserByUsername получает юзера по имени
func (us *AccessObject) GetUserByUsername(username string) (*UserModel, error) {
	u, err := us.getUserImpl(pqConn, "username", username)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrNotExists
		}

		return nil, errors.Wrapf(utils.ErrInternal, "user get by username error: %s", err.Error())
	}

	return u, nil
}

//nolint: gosec
func (us *AccessObject) getUserImpl(q postgresql.Queryer, field, value string) (*UserModel, error) {
	u := &UserModel{}

	row := q.QueryRow(`SELECT u.id, u.username, u.password,
	 					u.active, u.photo_uuid FROM users u WHERE `+field+` = $1;`, value)
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordCrypt, &u.Active, &u.PhotoUUID); err != nil {
		return nil, err
	}

	return u, nil
}

// GetUsersByIDs получает список юзеров по массиву ID
func (us *AccessObject) GetUsersByIDs(ids []int64) ([]*UserModel, error) {
	placeholders := make([]string, len(ids))
	for i, id := range ids {
		placeholders[i] = strconv.FormatInt(id, 10)
	}

	//nolint: gosec тут точно инты и никакие хакеры ничего не сломают
	rows, err := pqConn.Query(fmt.Sprintf(`SELECT u.id, u.username, u.password,
	 					u.active, u.photo_uuid FROM users u WHERE id IN (%s);`, strings.Join(placeholders, ",")))
	if err != nil {
		return nil, errors.Wrapf(utils.ErrInternal, "users get by ids error: %s", err.Error())
	}
	defer rows.Close()

	users := make([]*UserModel, 0)
	for rows.Next() {
		u := &UserModel{}
		err = rows.Scan(&u.ID, &u.Username,
			&u.PasswordCrypt, &u.Active,
			&u.PhotoUUID)
		if err != nil {
			return nil, errors.Wrapf(utils.ErrInternal, "users get by ids user scan error: %s", err.Error())
		}

		users = append(users, u)
	}

	return users, nil
}
