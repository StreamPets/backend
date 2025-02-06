package routes

import (
	"context"

	"github.com/google/uuid"
	"github.com/streampets/backend/announcers"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/services"
	"github.com/streampets/backend/twitch"
)

const ChannelId string = "channelId"
const OverlayId string = "overlayId"
const XExtensionJwt string = "x-extension-jwt"

type overlayIdGetter interface {
	GetOverlayId(channelId twitch.Id) (uuid.UUID, error)
}

type overlayIdValidator interface {
	ValidateOverlayId(channelId twitch.Id, overlayId uuid.UUID) error
}

type tokenValidator interface {
	ValidateToken(ctx context.Context, accessToken string) (twitch.Id, error)
}

type clientAdder interface {
	AddClient(channelId twitch.Id) announcers.Client
}

type clientRemover interface {
	RemoveClient(client announcers.Client)
}

type clientAddRemover interface {
	clientAdder
	clientRemover
}

type extTokenVerifier interface {
	VerifyExtToken(tokenString string) (*services.ExtToken, error)
}

type channelItemGetter interface {
	GetChannelsItems(channelId twitch.Id) ([]models.Item, error)
}
