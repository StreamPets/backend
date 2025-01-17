package services

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
)

var ErrIdMismatch = errors.New("channel id and overlay id do not match")
var ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
var ErrInvalidToken = errors.New("token is not valid")

type ExtToken struct {
	ChannelId models.TwitchId `json:"channel_id"`
	UserId    models.TwitchId `json:"user_id"`
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

type AuthService struct {
	channelRepo  OverlayIdGetter
	clientSecret string
}

type OverlayIdGetter interface {
	GetOverlayId(channelId models.TwitchId) (uuid.UUID, error)
}

func NewAuthService(
	channelRepo OverlayIdGetter,
	clientSecret string,
) *AuthService {
	return &AuthService{
		channelRepo:  channelRepo,
		clientSecret: clientSecret,
	}
}

func (s *AuthService) VerifyOverlayId(channelId models.TwitchId, overlayId uuid.UUID) error {
	expectedId, err := s.channelRepo.GetOverlayId(channelId)
	if err != nil {
		return err
	}

	if overlayId != expectedId {
		return ErrIdMismatch
	}

	return nil
}

func (s *AuthService) VerifyExtToken(tokenString string) (*ExtToken, error) {
	token, err := jwt.ParseWithClaims(tokenString, &ExtToken{}, s.keyFunc)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*ExtToken)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *AuthService) VerifyReceipt(tokenString string) (*Receipt, error) {
	fakeToken, err := jwt.Parse(tokenString, s.keyFunc)
	if err != nil {
		return nil, err
	}
	fmt.Println(fakeToken)

	token, err := jwt.ParseWithClaims(tokenString, &Receipt{}, s.keyFunc)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Receipt)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *AuthService) keyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, ErrUnexpectedSigningMethod
	}
	return []byte(s.clientSecret), nil
}
