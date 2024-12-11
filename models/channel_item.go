package models

import (
	"github.com/google/uuid"
)

type ChannelItem struct {
	ChannelID TwitchID
	ItemID    uuid.UUID
}
