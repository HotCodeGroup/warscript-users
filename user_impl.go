package main

import (
	"github.com/HotCodeGroup/warscript-utils/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/pgtype"
	"github.com/pkg/errors"
)

func getInfoUserByIDImpl(id int64) (*InfoUser, error) {
	user, err := Users.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	photoUUID := ""
	if user.PhotoUUID.Status == pgtype.Present {
		photoUUID = uuid.UUID(user.PhotoUUID.Bytes).String()
	}

	return &InfoUser{
		ID:     user.ID.Int,
		Active: user.Active.Bool,
		BasicUser: BasicUser{
			Username:  user.Username.String,
			PhotoUUID: photoUUID, // точно знаем, что там 16 байт
		},
	}, nil
}

//nolint: gocyclo
func updateUserImpl(info *SessionPayload, updateForm *FormUserUpdate) error {
	if err := updateForm.Validate(); err != nil {
		return err
	}

	// нечего обновлять
	if !updateForm.Username.IsDefined() &&
		!updateForm.NewPassword.IsDefined() &&
		!updateForm.PhotoUUID.IsDefined() {
		return nil
	}

	// взяли юзера
	user, err := Users.GetUserByID(info.ID)
	if err != nil {
		return errors.Wrap(err, "get user error")
	}

	// хотим обновить username
	if updateForm.Username.IsDefined() {
		user.Username = pgtype.Varchar{
			String: updateForm.Username.V,
			Status: pgtype.Present,
		}
	}

	if updateForm.PhotoUUID.IsDefined() {
		var photoUUID uuid.UUID
		status := pgtype.Null
		if updateForm.PhotoUUID.V != "" {
			status = pgtype.Present
			photoUUID = uuid.MustParse(updateForm.PhotoUUID.V)
		}

		user.PhotoUUID = pgtype.UUID{
			Bytes:  photoUUID,
			Status: status,
		}
	}

	// Если обновляется пароль, нужно проверить,
	// что пользователь знает старый
	if updateForm.NewPassword.IsDefined() {
		if !updateForm.OldPassword.IsDefined() {
			return &utils.ValidationError{
				"oldPassword": utils.ErrRequired.Error(),
			}
		}

		if !Users.CheckPassword(user, updateForm.OldPassword.V) {
			return &utils.ValidationError{
				"oldPassword": utils.ErrInvalid.Error(),
			}
		}

		user.Password = &updateForm.NewPassword.V
	}

	// пытаемся сохранить
	if err := Users.Save(user); err != nil {
		if errors.Cause(err) == utils.ErrTaken {
			return &utils.ValidationError{
				"username": utils.ErrTaken.Error(),
			}
		}

		return errors.Wrap(err, "user save error")
	}

	return nil
}
