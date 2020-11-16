package user

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
	domain_errors "go-ddd-cqrs-example/domain/errors"
)

// GetActive fetches an active user.
func GetActive(db gorm.DB, pk uuid.UUID, version *uint32) (*ActiveUser, error) {
	var user User

	err := db.Model(&user).Where("id = ?", pk).Take(&user).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, fmt.Errorf("User not found: %w", UserNotFound{})
	} else if user.IsActive == false {
		return nil, fmt.Errorf("Invariant failed: %w", IsInactive{})
	} else if version != nil && user.Version != *version {
		return nil, fmt.Errorf("Invalid version tag: %w", domain_errors.InvalidVersion{})
	} else if err != nil {
		return nil, fmt.Errorf("Error loading active user: %w", err)
	}

	return &ActiveUser{
		ID:           user.ID,
		EmailAddress: user.EmailAddress,
		Version:      user.Version,
	}, nil
}

// GetInactive fetches an inactive user.
func GetInactive(db gorm.DB, pk uuid.UUID, version *uint32) (*InactiveUser, error) {
	var user User

	err := db.Model(&user).Where("id = ?", pk).Take(&user).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, fmt.Errorf("User not found: %w", UserNotFound{})
	} else if user.IsActive == true {
		return nil, fmt.Errorf("Invariant failed: %w", IsActive{})
	} else if version != nil && user.Version != *version {
		return nil, fmt.Errorf("Invalid version tag: %w", domain_errors.InvalidVersion{})
	} else if err != nil {
		return nil, fmt.Errorf("Error loading inactive user: %w", err)
	}

	return &InactiveUser{
		ID:           user.ID,
		EmailAddress: user.EmailAddress,
		Version:      user.Version,
	}, err
}

// GetActiveByEmail fetches an active user by email for authentication when there's no token to extract user id from claims.
func GetActiveByEmail(db gorm.DB, emailAddress string, version *uint32) (*ActiveUser, error) {
	var user User

	err := db.Model(&user).Where("email_address = ?", emailAddress).Take(&user).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, fmt.Errorf("User not found: %w", UserNotFound{})
	} else if version != nil && user.Version != *version {
		return nil, fmt.Errorf("Invalid version tag: %w", domain_errors.InvalidVersion{})
	} else if user.IsActive == false {
		return nil, fmt.Errorf("Invariant failed: %w", IsInactive{})
	} else if err != nil {
		return nil, fmt.Errorf("Error loading active user: %w", err)
	}

	return &ActiveUser{
		ID:           user.ID,
		EmailAddress: user.EmailAddress,
		Version:      user.Version,
	}, nil
}

// GetUserPasswordHash to compare hashed with entered password.
func GetUserPasswordHash(db gorm.DB, email string, version *uint32) (*string, error) {
	var user User

	err := db.Model(user).Select("password").Where("email_address = ?", email).Take(&user).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, fmt.Errorf("User not found: %w", UserNotFound{})
	} else if version != nil && user.Version != *version {
		return nil, fmt.Errorf("Invalid version tag: %w", domain_errors.InvalidVersion{})
	} else if err != nil {
		return nil, fmt.Errorf("Error loading user password hash: %w", err)
	}

	return &user.Password, err
}
