package models

import "github.com/google/uuid"

type Rarity string

const (
	Common   Rarity = "common"
	Uncommon Rarity = "uncommon"
)

type Item struct {
	ItemID  uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	Name    string
	Rarity  Rarity `json:"rarity"`
	Image   string `json:"img"`
	PrevImg string `json:"prev"`
}
