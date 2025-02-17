package models

import "github.com/streampets/backend/twitch"

type User struct {
	UserId   twitch.UserId `gorm:"primaryKey"`
	Username string
}
