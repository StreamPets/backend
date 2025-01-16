package models

import (
	"github.com/google/uuid"
)

type ChannelItem struct {
	ChannelId UserId    `gorm:"primaryKey"`
	ItemId    uuid.UUID `gorm:"primaryKey;type:uuid"`
}
