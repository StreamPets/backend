package models

import "github.com/google/uuid"

type DefaultChannelItem struct {
	ChannelID TwitchID
	ItemID    uuid.UUID
}
