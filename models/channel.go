package models

import "github.com/google/uuid"

type Channel struct {
	ChannelID   string
	ChannelName string
	OverlayID   uuid.UUID
}
