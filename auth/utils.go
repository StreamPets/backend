package auth

import (
	"errors"
)

const XExtensionJwt = "x-extension-jwt"
const TransactionId = "transactionId"
const Rarity = "rarity"
const ChannelId = "channelId"
const UserId = "userId"
const OverlayId = "overlayId"

var ErrInvalidToken = errors.New("token is not valid")
var ErrIdMismatch = errors.New("channel id and overlay id do not match")
var ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
