package models

import (
	"github.com/google/uuid"
)

type OwnedItem struct {
	UserId        string    `gorm:"primaryKey"`
	ChannelId     string    `gorm:"primaryKey"`
	ItemId        uuid.UUID `gorm:"primaryKey;type:uuid"`
	TransactionId uuid.UUID `gorm:"unique"`
}
