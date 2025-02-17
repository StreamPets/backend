package gorm

import "github.com/google/uuid"

type channel struct {
	ChannelId   string `gorm:"primaryKey"`
	ChannelName string
	OverlayId   uuid.UUID `gorm:"type:uuid"`
}

type selectedItem struct {
	UserId    string    `gorm:"primaryKey"`
	ChannelId string    `gorm:"primaryKey"`
	ItemId    uuid.UUID `gorm:"type:uuid"`
}

type ownedItem struct {
	UserId        string    `gorm:"primaryKey"`
	ChannelId     string    `gorm:"primaryKey"`
	ItemId        uuid.UUID `gorm:"primaryKey;type:uuid"`
	TransactionId uuid.UUID `gorm:"unique"`
}

type channelItem struct {
	ChannelId string    `gorm:"primaryKey"`
	ItemId    uuid.UUID `gorm:"primaryKey;type:uuid"`
}

type defaultChannelItem struct {
	ChannelId string    `gorm:"primaryKey"`
	ItemId    uuid.UUID `gorm:"type:uuid"`
}
