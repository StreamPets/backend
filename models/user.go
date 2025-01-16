package models

type UserId string

type Viewer struct {
	ViewerId UserId `gorm:"primaryKey"`
	Username string
}
