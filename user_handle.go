package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/HotCodeGroup/warscript-utils/utils"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// CheckUsername checks if username already used
func CheckUsername(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger(r, logger, "CheckUsername")
	errWriter := utils.NewErrorResponseWriter(w, logger)

	bUser := &BasicUser{}
	err := utils.DecodeBodyJSON(r.Body, bUser)
	if err != nil {
		errWriter.WriteWarn(http.StatusBadRequest, errors.Wrap(err, "decode body error"))
		return
	}

	_, err = Users.GetUserByUsername(bUser.Username) // если база лежит
	if err != nil && errors.Cause(err) != utils.ErrNotExists {
		errWriter.WriteError(http.StatusInternalServerError, errors.Wrap(err, "get user method error"))
		return
	}

	// Если нет ошибки, то такой юзер точно есть, а если ошибка есть,
	// то это точно models.ErrNotExists,
	// так как остальные вышли бы раньше
	used := (err == nil)
	utils.WriteApplicationJSON(w, http.StatusOK, &struct {
		Used bool `json:"used"`
	}{
		Used: used,
	})
}

// GetUser get user info by ID
func GetUser(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger(r, logger, "GetUser")
	errWriter := utils.NewErrorResponseWriter(w, logger)
	vars := mux.Vars(r)

	userID, err := strconv.ParseInt(vars["user_id"], 10, 64)
	if err != nil {
		errWriter.WriteError(http.StatusNotFound, errors.Wrap(err, "wrong format user_id"))
		return
	}

	infoUser, err := getInfoUserByIDImpl(userID)
	if err != nil {
		if errors.Cause(err) == utils.ErrNotExists {
			errWriter.WriteWarn(http.StatusNotFound, errors.Wrap(err, "user not exists"))
		} else {
			errWriter.WriteError(http.StatusInternalServerError, errors.Wrap(err, "get user method error"))
		}
		return
	}

	// !!! отдаём только ту часть, которая без секрета
	utils.WriteApplicationJSON(w, http.StatusOK, infoUser.InfoUser)
}

// UpdateUser обновляет данные пользователя
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger(r, logger, "UpdateUser")
	errWriter := utils.NewErrorResponseWriter(w, logger)
	info := SessionInfo(r)
	if info == nil {
		errWriter.WriteWarn(http.StatusUnauthorized, errors.New("session info is not presented"))
		return
	}

	updateForm := &FormUserUpdate{}
	err := utils.DecodeBodyJSON(r.Body, updateForm)
	if err != nil {
		errWriter.WriteWarn(http.StatusBadRequest, errors.Wrap(err, "decode body error"))
		return
	}

	err = updateUserImpl(info, updateForm)
	if err != nil {
		if validErr, ok := err.(*utils.ValidationError); ok {
			errWriter.WriteValidationError(validErr)
			return
		}

		if errors.Cause(err) == utils.ErrNotExists {
			errWriter.WriteWarn(http.StatusUnauthorized, errors.Wrap(err, "user not exists"))
		} else {
			errWriter.WriteError(http.StatusInternalServerError, err)
		}

		return
	}

	w.WriteHeader(http.StatusOK)
}

// CreateUser creates new user
func CreateUser(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger(r, logger, "CreateUser")
	errWriter := utils.NewErrorResponseWriter(w, logger)

	form := &FormUser{}
	err := utils.DecodeBodyJSON(r.Body, form)
	if err != nil {
		errWriter.WriteWarn(http.StatusBadRequest, errors.Wrap(err, "decode body error"))
		return
	}

	if valError := form.Validate(); valError != nil {
		errWriter.WriteValidationError(valError)
		return
	}

	user := &UserModel{
		Username: form.Username,
		Password: &form.Password,
	}

	if err = Users.Create(user); err != nil {
		if errors.Cause(err) == utils.ErrTaken {
			errWriter.WriteValidationError(&utils.ValidationError{
				"username": utils.ErrTaken.Error(),
			})
		} else {
			errWriter.WriteError(http.StatusInternalServerError, errors.Wrap(err, "user create error"))
		}
		return
	}

	// сразу же логиним юзера
	session, err := createSessionImpl(form)
	if err != nil {
		if validErr, ok := err.(*utils.ValidationError); ok {
			errWriter.WriteValidationError(validErr)
			return
		}

		errWriter.WriteError(http.StatusInternalServerError, err)
		return
	}

	// ставим куку
	http.SetCookie(w, &http.Cookie{
		Name:     "JSESSIONID",
		Path:     "/",
		Value:    session.Token,
		Expires:  time.Now().Add(2628000 * time.Second),
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
}
