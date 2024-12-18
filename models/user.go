package models

type TwitchID string

type User struct {
	UserID   TwitchID `gorm:"primaryKey"`
	Username string
}
