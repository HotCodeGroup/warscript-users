package main

import (
	"database/sql"

	"github.com/HotCodeGroup/warscript-utils/models"
	"github.com/HotCodeGroup/warscript-utils/utils"
	"github.com/pkg/errors"
)

func getInfoUserByIDImpl(id int64) (*ProfileInfoUser, error) {
	user, err := Users.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	return &ProfileInfoUser{
		InfoUser: InfoUser{
			ID:     user.ID,
			Active: user.Active,
			BasicUser: BasicUser{
				Username:  user.Username,
				PhotoUUID: user.GetPhotoUUID(), // точно знаем, что там 16 байт
			},
		},
		VkSecret: user.VkSecret,
	}, nil
}

//nolint: gocyclo
func updateUserImpl(info *models.SessionPayload, updateForm *FormUserUpdate) error {
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
		user.Username = updateForm.Username.V
	}

	if updateForm.PhotoUUID.IsDefined() {
		user.PhotoUUID = sql.NullString{String: updateForm.PhotoUUID.V, Valid: true}
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
