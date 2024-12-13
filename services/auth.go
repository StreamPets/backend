package services

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/repositories"
)

var ErrIdMismatch = errors.New("channelID and overlayID do not match")
var ErrUnexpectedSigningMethod = errors.New("unexpected signing method")
var ErrInvalidToken = errors.New("token is not valid")

type ExtToken struct {
	ChannelID models.TwitchID `json:"channel_id"`
	UserID    models.TwitchID `json:"user_id"`
	jwt.RegisteredClaims
}

type Receipt struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	jwt.RegisteredClaims
}

type AuthService interface {
	VerifyOverlayID(channelID models.TwitchID, overlayID uuid.UUID) error
	VerifyExtToken(tokenString string) (*ExtToken, error)
	VerifyReceipt(tokenString string) (*Receipt, error)
}

type authService struct {
	channelRepo  repositories.ChannelRepo
	clientSecret string
}

func NewAuthService(
	channelRepo repositories.ChannelRepo,
	clientSecret string,
) AuthService {
	return &authService{
		channelRepo:  channelRepo,
		clientSecret: clientSecret,
	}
}

func (s *authService) VerifyOverlayID(channelID models.TwitchID, overlayID uuid.UUID) error {
	expectedID, err := s.channelRepo.GetOverlayID(channelID)
	if err != nil {
		return err
	}

	if overlayID != expectedID {
		return ErrIdMismatch
	}

	return nil
}

func (s *authService) VerifyExtToken(tokenString string) (*ExtToken, error) {
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

func (s *authService) VerifyReceipt(tokenString string) (*Receipt, error) {
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

func (s *authService) keyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, ErrUnexpectedSigningMethod
	}
	return []byte(s.clientSecret), nil
}
