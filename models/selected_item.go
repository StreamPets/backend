package models

import (
	"github.com/google/uuid"
)

type SelectedItem struct {
	UserID    TwitchID `gorm:"primaryKey"`
	ChannelID TwitchID `gorm:"primaryKey"`
	ItemID    uuid.UUID
}
