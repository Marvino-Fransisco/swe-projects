package user

import (
	"time"

	"gorm.io/gorm"

	"shared/domain/shared"
)

// User represents a registered user in the system.
type User struct {
	ID           string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FullName     FullName  `gorm:"type:varchar(255);not null" json:"full_name"`
	Email        Email     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Address      string    `gorm:"type:text" json:"address"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// BeforeCreate hook to generate UUID if not set.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		id, err := shared.GenerateUUID()
		if err != nil {
			return err
		}
		u.ID = id
	}
	return nil
}
