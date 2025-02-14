package auth

import (
	"errors"
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

		token, err := s.verifyExtToken(tokenString)
		if errors.As(err, &ErrInvalidToken{}) {
			slog.Debug("invalid access token in header")
			ctx.JSON(http.StatusUnauthorized, nil)
			return
		} else if err != nil {
			slog.Error("error when validating access token", "err", err.Error())
			ctx.JSON(http.StatusInternalServerError, nil)
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
		err := ctx.ShouldBindJSON(request)
		if err != nil {
			slog.Warn("failed to bind json")
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}

		receipt, err := s.VerifyReceipt(request.Receipt)
		e := new(ErrInvalidToken)
		if errors.As(err, e) {
			slog.Warn("invalid token", "token", e.TokenString)
			ctx.JSON(http.StatusUnauthorized, nil)
			return
		} else if err != nil {
			slog.Error("failed to validate token", "err", err.Error())
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		ctx.Set(TransactionId, receipt.Data.TransactionId)
		ctx.Set(Rarity, receipt.Data.Product.Rarity)
		ctx.Next()
	}
}

// TODO: Handle other errors
func (s *AuthService) verifyExtToken(tokenString string) (*extToken, error) {
	token, err := jwt.ParseWithClaims(tokenString, &extToken{}, s.keyFunc)
	if err != nil {
		return nil, NewErrInvalidToken(tokenString)
	}

	claims, ok := token.Claims.(*extToken)
	if !ok || !token.Valid {
		// TODO: When can this line be reached?
		slog.Error("process reached area it should not have been able to", "token", token)
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
