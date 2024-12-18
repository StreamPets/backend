package models

import "github.com/google/uuid"

type OwnedItem struct {
	UserID        TwitchID  `gorm:"primaryKey"`
	ChannelID     TwitchID  `gorm:"primaryKey"`
	ItemID        uuid.UUID `gorm:"primaryKey;type:uuid"`
	TransactionID uuid.UUID `gorm:"unique"`
}
