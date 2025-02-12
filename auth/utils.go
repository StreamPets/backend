package auth

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/twitch"
)

var ErrIdMismatch = errors.New("channel id and overlay id do not match")
var ErrUnexpectedSigningMethod = errors.New("unexpected signing method")

type ErrInvalidToken struct {
	TokenString string
}

func NewErrInvalidToken(tokenString string) *ErrInvalidToken {
	return &ErrInvalidToken{TokenString: tokenString}
}

func (e *ErrInvalidToken) Error() string {
	return "token is not valid"
}

type ExtToken struct {
	ChannelId twitch.Id `json:"channel_id"`
	UserId    twitch.Id `json:"user_id"`
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
