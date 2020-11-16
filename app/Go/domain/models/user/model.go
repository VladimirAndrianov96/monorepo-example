package user

import (
	"errors"
	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
	"time"
)

// PendingUser represents a user about to sign up.
type PendingUser struct {
	ID           uuid.UUID
	EmailAddress string
	Password     string
}

// ActiveUser represents an active user in the system.
type ActiveUser struct {
	ID           uuid.UUID
	EmailAddress string
	Version      uint32
}

// InactiveUser represents a deactivated user in the system.
type InactiveUser struct {
	ID           uuid.UUID
	EmailAddress string
	Version      uint32
}

// User represents a persistence model for the user entity.
type User struct {
	ID           uuid.UUID `gorm:"primary_key;index:idx_member" json:"id"`
	EmailAddress string    `gorm:"not null;unique;index:idx_member" json:"email_address"`
	Password     string    `gorm:"not null;unique" json:"password"`
	IsActive     bool      `gorm:"not null" json:"is_active"`
	CreatedAt    time.Time `gorm:"default:now();not null" json:"created_at"`
	Version      uint32    `gorm:"not null" json:"version"`
}

// VerifyUserPassword with the hash stored in database.
func VerifyUserPassword(db *gorm.DB, emailAddress string, password string) error {
	userPasswordHash, err := GetUserPasswordHash(*db, emailAddress, nil)
	if err != nil {
		return err
	}

	return bcrypt.CompareHashAndPassword([]byte(*userPasswordHash), []byte(password))
}

// Hash the password.
func Hash(password string) (*string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	hashString := string(hash)

	return &hashString, err
}

func isEmailAddressUnique(db gorm.DB, emailAddress string) (error, bool) {
	var user User

	if err := db.Model(&user).Where(
		"email_address = ?",
		emailAddress,
	).First(&user).Error; err != nil {
		if errors.As(err, &gorm.ErrRecordNotFound) {
			return nil, true
		}
		return err, false
	}

	return nil, false
}
