package models

import "github.com/google/uuid"

type DefaultChannelItem struct {
	ChannelID string
	ItemID    uuid.UUID
}
