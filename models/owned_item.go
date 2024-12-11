package models

import "github.com/google/uuid"

type OwnedItem struct {
	UserID        TwitchID
	ChannelID     TwitchID
	ItemID        uuid.UUID
	TransactionID uuid.UUID
}
