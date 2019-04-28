package main

import (
	"net/http"
	"time"

	"github.com/HotCodeGroup/warscript-utils/middlewares"
	"github.com/HotCodeGroup/warscript-utils/models"
	"github.com/HotCodeGroup/warscript-utils/utils"

	"github.com/pkg/errors"
)

// SessionInfo достаёт инфу о юзере из контекстаs
func SessionInfo(r *http.Request) *SessionPayload {
	if rv := r.Context().Value(middlewares.SessionInfoKey); rv != nil {
		if rInfo, ok := rv.(*models.SessionPayload); ok {
			return &SessionPayload{rInfo.ID}
		}
	}

	return nil
}

// CreateSession вход + кука
func CreateSession(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger(r, logger, "SignInUser")
	errWriter := utils.NewErrorResponseWriter(w, logger)

	form := &FormUser{}
	err := utils.DecodeBodyJSON(r.Body, form)
	if err != nil {
		errWriter.WriteWarn(http.StatusBadRequest, errors.Wrap(err, "decode body error"))
		return
	}

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

// DeleteSession выход + удаление куки
func DeleteSession(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger(r, logger, "SignOutUser")
	errWriter := utils.NewErrorResponseWriter(w, logger)

	cookie, err := r.Cookie("JSESSIONID")
	if err != nil {
		errWriter.WriteWarn(http.StatusUnauthorized, errors.Wrap(err, "get cookie error"))
		return
	}

	session := &Session{
		Token: cookie.Value,
	}
	err = Sessions.Delete(session)
	if err != nil {
		errWriter.WriteWarn(http.StatusInternalServerError, errors.Wrap(err, "session delete error"))
		return
	}

	cookie.Expires = time.Unix(0, 0)
	http.SetCookie(w, cookie)

	w.WriteHeader(http.StatusOK)
}

// GetSession возвращает сессмю
func GetSession(w http.ResponseWriter, r *http.Request) {
	logger := utils.GetLogger(r, logger, "GetSession")
	errWriter := utils.NewErrorResponseWriter(w, logger)
	info := SessionInfo(r)
	if info == nil {
		errWriter.WriteWarn(http.StatusUnauthorized, errors.New("session info is not presented"))
		return
	}

	infoUser, err := getInfoUserByIDImpl(info.ID)
	if err != nil {
		if errors.Cause(err) == utils.ErrNotExists {
			errWriter.WriteWarn(http.StatusUnauthorized, errors.Wrap(err, "user not exists"))
		} else {
			errWriter.WriteError(http.StatusInternalServerError, err)
		}
		return
	}

	utils.WriteApplicationJSON(w, http.StatusOK, infoUser)
}
