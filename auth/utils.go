package auth

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
)

const XExtensionJwt = "x-extension-jwt"
const TransactionId = "transactionId"
const Rarity = "rarity"
const ChannelId = "channelId"
const UserId = "userId"

var ErrInvalidToken = errors.New("token is not valid")
var ErrIdMismatch = errors.New("channel id and overlay id do not match")
var ErrUnexpectedSigningMethod = errors.New("unexpected signing method")

type extToken struct {
	ChannelId string `json:"channel_id"`
	UserId    string `json:"user_id"`
	jwt.RegisteredClaims
}

type Product struct {
	Rarity models.Rarity `json:"sku"`
}

type Data struct {
	TransactionId uuid.UUID `json:"transactionId"`
	Product       Product   `json:"product"`
}

type Receipt struct {
	Data Data `json:"data"`
	jwt.RegisteredClaims
}
