package models

import (
	"github.com/google/uuid"
)

type ChannelItem struct {
	ChannelID TwitchID  `gorm:"primaryKey"`
	ItemID    uuid.UUID `gorm:"primaryKey;type:uuid"`
}
