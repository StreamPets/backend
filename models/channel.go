package models

import "github.com/google/uuid"

type Channel struct {
	ChannelId   UserId `gorm:"primaryKey"`
	ChannelName string
	OverlayId   uuid.UUID `gorm:"type:uuid"`
}
