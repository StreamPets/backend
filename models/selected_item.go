package models

import (
	"github.com/google/uuid"
	"github.com/streampets/backend/twitch"
)

type SelectedItem struct {
	UserId    twitch.UserId `gorm:"primaryKey"`
	ChannelId twitch.UserId `gorm:"primaryKey"`
	ItemId    uuid.UUID     `gorm:"type:uuid"`
}
