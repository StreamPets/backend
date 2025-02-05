package models

import "github.com/streampets/backend/twitch"

type User struct {
	UserId   twitch.Id `gorm:"primaryKey"`
	Username string
}
