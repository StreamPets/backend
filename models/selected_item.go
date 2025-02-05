package models

import (
	"github.com/google/uuid"
	"github.com/streampets/backend/twitch"
)

type SelectedItem struct {
	UserId    twitch.Id `gorm:"primaryKey"`
	ChannelId twitch.Id `gorm:"primaryKey"`
	ItemId    uuid.UUID `gorm:"type:uuid"`
}
