package auth

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/streampets/backend/twitch"
)

type AuthService struct {
	getOverlayId func(channelId twitch.Id) (uuid.UUID, error)
	clientSecret string
}

func New(
	getOverlayId func(channelId twitch.Id) (uuid.UUID, error),
	clientSecret string,
) *AuthService {
	return &AuthService{
		getOverlayId: getOverlayId,
		clientSecret: clientSecret,
	}
}

func (s *AuthService) ValidateOverlayId(channelId twitch.Id, overlayId uuid.UUID) error {
	expectedId, err := s.getOverlayId(channelId)
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
		return nil, NewErrInvalidToken(tokenString)
	}

	return claims, nil
}

func (s *AuthService) VerifyReceipt(tokenString string) (*Receipt, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Receipt{}, s.keyFunc)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Receipt)
	if !ok || !token.Valid {
		return nil, NewErrInvalidToken(tokenString)
	}

	return claims, nil
}

func (s *AuthService) keyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, ErrUnexpectedSigningMethod
	}
	return []byte(s.clientSecret), nil
}
