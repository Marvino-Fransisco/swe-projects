package user

import "time"

type UserPreferences struct {
	UserID    string    `gorm:"type:uuid;primaryKey" json:"user_id"`
	Theme     string    `gorm:"type:varchar(100)" json:"theme"`
	Language  string    `gorm:"type:varchar(100)" json:"language"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}
