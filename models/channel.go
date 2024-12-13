package models

import "github.com/google/uuid"

type Channel struct {
	ChannelID   TwitchID `gorm:"primaryKey"`
	ChannelName string
	OverlayID   uuid.UUID `gorm:"type:uuid"`
}
