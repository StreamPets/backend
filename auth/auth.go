package auth

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
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

func (s *AuthService) ExtensionMiddleware() func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		tokenString := ctx.GetHeader(XExtensionJwt)

		token := new(extToken)
		if err := s.verifyToken(tokenString, token); err != nil {
			slog.Error("error when validating access token", "err", err.Error())
			ctx.JSON(http.StatusUnauthorized, nil)
			return
		}

		ctx.Set(ChannelId, token.ChannelId)
		ctx.Set(UserId, token.UserId)
		ctx.Next()
	}
}

func (s *AuthService) ReceiptMiddleware() func(ctx *gin.Context) {

	type request struct {
		Receipt string `json:"receipt"`
	}

	return func(ctx *gin.Context) {
		request := new(request)
		if err := ctx.ShouldBindJSON(request); err != nil {
			slog.Error("failed to bind json: no 'receipt' field")
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}

		receipt := new(Receipt)
		if err := s.verifyToken(request.Receipt, receipt); err != nil {
			slog.Error("error when validating receipt", "err", err.Error())
			ctx.JSON(http.StatusUnauthorized, nil)
			return
		}

		ctx.Set(TransactionId, receipt.Data.TransactionId)
		ctx.Set(Rarity, receipt.Data.Product.Rarity)
		ctx.Next()
	}
}

func (s *AuthService) verifyToken(tokenString string, claims jwt.Claims) error {
	token, err := jwt.ParseWithClaims(tokenString, claims, s.keyFunc)
	if err != nil {
		return err
	}

	if !token.Valid {
		return ErrInvalidToken
	}

	return nil
}

func (s *AuthService) keyFunc(token *jwt.Token) (interface{}, error) {
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, ErrUnexpectedSigningMethod
	}
	return []byte(s.clientSecret), nil
}
