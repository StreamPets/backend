package models

import (
	"github.com/google/uuid"
)

type SelectedItem struct {
	ViewerId  UserId    `gorm:"primaryKey"`
	ChannelId UserId    `gorm:"primaryKey"`
	ItemId    uuid.UUID `gorm:"type:uuid"`
}
