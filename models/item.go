package models

import "github.com/google/uuid"

type Rarity string

const (
	Common   Rarity = "common"
	Uncommon Rarity = "uncommon"
)

type Item struct {
	ItemID  uuid.UUID `gorm:"primaryKey;type:uuid"`
	Name    string
	Rarity  Rarity
	Image   string
	PrevImg string
}
