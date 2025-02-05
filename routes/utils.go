package routes

import (
	"context"

	"github.com/google/uuid"
	"github.com/streampets/backend/twitch"
)

type OverlayIdGetter interface {
	GetOverlayId(channelId twitch.Id) (uuid.UUID, error)
}

type TokenValidator interface {
	ValidateToken(ctx context.Context, accessToken string) (twitch.Id, error)
}
