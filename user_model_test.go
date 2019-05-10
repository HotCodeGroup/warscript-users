package main

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/HotCodeGroup/warscript-utils/utils"
	"github.com/pkg/errors"
)

func TestCreateOK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT").WillReturnError(sql.ErrNoRows)
	mock.ExpectExec("INSERT INTO users").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	pqConn = db
	Users = &AccessObject{}

	pass := "lol"
	u := &UserModel{
		Username: "kek",
		Password: &pass,
	}

	if err = Users.Create(u); err != nil {
		t.Errorf("TestCreate got unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestCreate there were unfulfilled expectations: %s", err)
	}
}

func TestCreateTaken(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "active", "photo_uuid"}).
			AddRow(1, "kek", []byte{1, 2, 3}, true, "kek"))
	mock.ExpectRollback()

	pqConn = db
	Users = &AccessObject{}

	pass := "lol"
	u := &UserModel{
		Username: "kek",
		Password: &pass,
	}

	if err = Users.Create(u); err != nil {
		if errors.Cause(err) != utils.ErrTaken {
			t.Errorf("TestCreate got unexpected error: %v", err)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestCreate there were unfulfilled expectations: %s", err)
	}
}

func TestCreateBeginErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	pqConn = db
	Users = &AccessObject{}

	pass := "lol"
	u := &UserModel{
		Username: "kek",
		Password: &pass,
	}

	if err = Users.Create(u); err != nil {
		if errors.Cause(err) != utils.ErrInternal {
			t.Errorf("TestCreateBeginErr got unexpected error: %v", err)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestCreateBeginErr there were unfulfilled expectations: %s", err)
	}
}

func TestSaveOK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT").WillReturnError(sql.ErrNoRows)
	mock.ExpectExec("UPDATE users").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	pqConn = db
	Users = &AccessObject{}

	pass := "lol"
	u := &UserModel{
		ID:       1,
		Username: "kek",
		Password: &pass,
		Active:   true,
	}

	if err = Users.Save(u); err != nil {
		t.Errorf("TestCreate got unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestCreate there were unfulfilled expectations: %s", err)
	}
}

func TestSaveTaken(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "active", "photo_uuid"}).
			AddRow(2, "kek", []byte{1, 2, 3}, true, "kek"))
	mock.ExpectRollback()

	pqConn = db
	Users = &AccessObject{}

	pass := "lol"
	u := &UserModel{
		ID:       1,
		Username: "kek",
		Password: &pass,
		Active:   true,
	}

	if err = Users.Save(u); err != nil {
		if errors.Cause(err) != utils.ErrTaken {
			t.Errorf("TestSaveTaken got unexpected error: %v, expected: %v", err, utils.ErrTaken)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestSaveTaken there were unfulfilled expectations: %s", err)
	}
}

func TestSaveBeginErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	pqConn = db
	Users = &AccessObject{}

	pass := "lol"
	u := &UserModel{
		Username: "kek",
		Password: &pass,
	}

	if err = Users.Save(u); err != nil {
		if errors.Cause(err) != utils.ErrInternal {
			t.Errorf("TestSaveBeginErr got unexpected error: %v", err)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestSaveBeginErr there were unfulfilled expectations: %s", err)
	}
}

func TestGetUsersByIDsModelOK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "active", "photo_uuid"}).
			AddRow(1, "kek1", []byte{1, 2, 3}, true, "kek").
			AddRow(2, "kek2", []byte{1, 2, 3}, true, "kek").
			AddRow(3, "kek3", []byte{1, 2, 3}, true, "kek"))

	pqConn = db
	Users = &AccessObject{}

	if _, err = Users.GetUsersByIDs([]int64{1, 2, 3}); err != nil {
		t.Errorf("TestCreate got unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestCreate there were unfulfilled expectations: %s", err)
	}
}

func TestGetUserByIDModelOK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "active", "photo_uuid"}).
			AddRow(1, "kek1", []byte{1, 2, 3}, true, "kek"))

	pqConn = db
	Users = &AccessObject{}

	if _, err = Users.GetUserByID(1); err != nil {
		t.Errorf("TestCreate got unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestCreate there were unfulfilled expectations: %s", err)
	}
}

func TestGetUserByIDModelNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT").WillReturnError(sql.ErrNoRows)

	pqConn = db
	Users = &AccessObject{}

	if _, err = Users.GetUserByID(1); err != nil {
		if errors.Cause(err) != utils.ErrNotExists {
			t.Errorf("TestCreate got unexpected error: %v", err)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestCreate there were unfulfilled expectations: %s", err)
	}
}

func TestGetUserByUsernameModelOK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT").
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password", "active", "photo_uuid"}).
			AddRow(1, "kek1", []byte{1, 2, 3}, true, "kek"))

	pqConn = db
	Users = &AccessObject{}

	if _, err = Users.GetUserByUsername("kek"); err != nil {
		t.Errorf("TestCreate got unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestCreate there were unfulfilled expectations: %s", err)
	}
}

func TestGetUserByUsernameModelNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT").WillReturnError(sql.ErrNoRows)

	pqConn = db
	Users = &AccessObject{}

	if _, err = Users.GetUserByUsername("kek"); err != nil {
		if errors.Cause(err) != utils.ErrNotExists {
			t.Errorf("TestCreate got unexpected error: %v", err)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("TestCreate there were unfulfilled expectations: %s", err)
	}
}
