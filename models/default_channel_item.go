package models

import "github.com/google/uuid"

type DefaultChannelItem struct {
	ChannelId UserId    `gorm:"primaryKey"`
	ItemId    uuid.UUID `gorm:"type:uuid"`
}
