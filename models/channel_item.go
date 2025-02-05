package models

import (
	"github.com/google/uuid"
	"github.com/streampets/backend/twitch"
)

type ChannelItem struct {
	ChannelId twitch.Id `gorm:"primaryKey"`
	ItemId    uuid.UUID `gorm:"primaryKey;type:uuid"`
}
