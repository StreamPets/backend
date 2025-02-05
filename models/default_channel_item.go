package models

import (
	"github.com/google/uuid"
	"github.com/streampets/backend/twitch"
)

type DefaultChannelItem struct {
	ChannelId twitch.Id `gorm:"primaryKey"`
	ItemId    uuid.UUID `gorm:"type:uuid"`
}
