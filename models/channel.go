package models

import (
	"github.com/google/uuid"
	"github.com/streampets/backend/twitch"
)

type Channel struct {
	ChannelId   twitch.Id `gorm:"primaryKey"`
	ChannelName string
	OverlayId   uuid.UUID `gorm:"type:uuid"`
}
