package models

import "github.com/google/uuid"

type DefaultChannelItem struct {
	ChannelID TwitchID  `gorm:"primaryKey"`
	ItemID    uuid.UUID `gorm:"type:uuid"`
}
