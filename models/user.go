package models

type TwitchId string

type Viewer struct {
	ViewerId TwitchId `gorm:"primaryKey"`
	Username string
}
