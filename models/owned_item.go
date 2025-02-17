package models

import (
	"github.com/google/uuid"
	"github.com/streampets/backend/twitch"
)

type OwnedItem struct {
	UserId        twitch.UserId `gorm:"primaryKey"`
	ChannelId     twitch.UserId `gorm:"primaryKey"`
	ItemId        uuid.UUID     `gorm:"primaryKey;type:uuid"`
	TransactionId uuid.UUID     `gorm:"unique"`
}
