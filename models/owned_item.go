package models

import "github.com/google/uuid"

type OwnedItem struct {
	ViewerId      TwitchId  `gorm:"primaryKey"`
	ChannelId     TwitchId  `gorm:"primaryKey"`
	ItemId        uuid.UUID `gorm:"primaryKey;type:uuid"`
	TransactionId uuid.UUID `gorm:"unique"`
}
