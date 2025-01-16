package models

import (
	"github.com/google/uuid"
)

type SelectedItem struct {
	ViewerId  TwitchId  `gorm:"primaryKey"`
	ChannelId TwitchId  `gorm:"primaryKey"`
	ItemId    uuid.UUID `gorm:"type:uuid"`
}
