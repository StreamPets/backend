package models

type TwitchId string

type User struct {
	UserId   TwitchId `gorm:"primaryKey"`
	Username string
}
