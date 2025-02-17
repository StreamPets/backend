package models

type User struct {
	UserId   string `gorm:"primaryKey"`
	Username string
}
