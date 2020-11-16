package user

import (
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/jinzhu/gorm"
	domain_errors "go-ddd-cqrs-example/domain/errors"
	"strings"
)

// Create a new active user.
func Create(db gorm.DB, pendingUser PendingUser) (*UserCreated, error) {
	pendingUser.EmailAddress = strings.TrimSpace(pendingUser.EmailAddress)
	pendingUser.EmailAddress = strings.ToLower(pendingUser.EmailAddress)

	if err := validation.ValidateStruct(
		&pendingUser,
		validation.Field(&pendingUser.ID, is.UUIDv4),
		validation.Field(&pendingUser.EmailAddress, validation.Required, is.Email),
		validation.Field(&pendingUser.Password, validation.Required, validation.Length(6, 20)),
	); err != nil {
		return nil, err
	}

	err, exists := isEmailAddressUnique(db, pendingUser.EmailAddress)
	if err != nil {
		return nil, err
	} else if exists != true {
		return nil, AlreadyExists{}
	}

	passwordHash, err := Hash(pendingUser.Password)
	if err != nil {
		return nil, err
	}

	activeUser := ActiveUser{
		ID:           pendingUser.ID,
		EmailAddress: pendingUser.EmailAddress,
		Version:      1,
	}

	if err := db.Create(&User{
		ID:           activeUser.ID,
		EmailAddress: activeUser.EmailAddress,
		Password:     *passwordHash,
		IsActive:     true,
		Version:      activeUser.Version,
	}).Error; err != nil {
		return nil, err
	}

	return &UserCreated{
		UserID:       activeUser.ID.String(),
		EmailAddress: activeUser.EmailAddress,
		Version:      activeUser.Version,
	}, nil
}

// Deactivate an active user.
func Deactivate(db gorm.DB, activeUser ActiveUser) (*UserDeactivated, error) {
	inactiveUser := InactiveUser{
		ID:      activeUser.ID,
		Version: activeUser.Version + 1,
	}

	// Update attributes with `struct`, will only update non-zero fields.
	// Update attributes with `map` instead.
	// https://gorm.io/docs/update.html#Updates-multiple-columns
	result := db.Model(&User{}).
		Where("id = ? AND version = ?",
			activeUser.ID,
			activeUser.Version,
		).Updates(map[string]interface{}{"is_active": false, "version": inactiveUser.Version})

	if result.Error != nil {
		return nil, fmt.Errorf("Error deactivating active user: %w", result.Error)
	} else if result.RowsAffected != 1 {
		return nil, fmt.Errorf("State conflict: %w", domain_errors.StateConflict{})
	}

	return &UserDeactivated{
		UserID:  inactiveUser.ID.String(),
		Version: inactiveUser.Version,
	}, nil
}

// Activate an inactive user.
func Activate(db gorm.DB, inactiveUser InactiveUser) (*UserActivated, error) {
	activeUser := ActiveUser{
		ID:      inactiveUser.ID,
		Version: inactiveUser.Version + 1,
	}

	// Update attributes with `struct`, will only update non-zero fields.
	// Update attributes with `map` instead.
	// https://gorm.io/docs/update.html#Updates-multiple-columns
	result := db.Model(&User{}).
		Where("id = ? AND version = ?",
			inactiveUser.ID,
			inactiveUser.Version,
		).Updates(map[string]interface{}{"is_active": true, "version": activeUser.Version})

	if result.Error != nil {
		return nil, fmt.Errorf("Error activating inactive user: %w", result.Error)
	} else if result.RowsAffected != 1 {
		return nil, fmt.Errorf("State conflict: %w", domain_errors.StateConflict{})
	}

	return &UserActivated{
		UserID:  activeUser.ID.String(),
		Version: activeUser.Version,
	}, nil
}
