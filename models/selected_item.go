package models

import (
	"github.com/google/uuid"
)

type SelectedItem struct {
	UserId    string    `gorm:"primaryKey"`
	ChannelId string    `gorm:"primaryKey"`
	ItemId    uuid.UUID `gorm:"type:uuid"`
}
