package models

import (
	"github.com/google/uuid"
)

type Channel struct {
	ChannelId   string `gorm:"primaryKey"`
	ChannelName string
	OverlayId   uuid.UUID `gorm:"type:uuid"`
}
