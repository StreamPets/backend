package auth

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	streampets "github.com/streampets/backend"
)

type AuthService struct {
	clientSecret string
}

func New(
	clientSecret string,
) *AuthService {
	return &AuthService{
		clientSecret: clientSecret,
	}
}

func ListenAuthentication(
	getOverlayId func(channelId string) (uuid.UUID, error),
) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		channelId := ctx.Query(ChannelId)
		overlayId, err := uuid.Parse(ctx.Query(OverlayId))
		if err != nil {
			slog.Debug("param is not uuid type")
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}

		if err = validateOverlayId(channelId, overlayId, getOverlayId); err != nil {
			slog.Warn("unrecognised overlay id and channel id pair", "channel id", channelId, "overlay id", overlayId)
			ctx.JSON(http.StatusUnauthorized, nil)
			return
		}

		ctx.Set(ChannelId, channelId)
		ctx.Next()
	}
}

func validateOverlayId(
	channelId string,
	overlayId uuid.UUID,
	getOverlayId func(channelId string) (uuid.UUID, error),
) error {
	expectedId, err := getOverlayId(channelId)
	if err != nil {
		return err
	}

	if overlayId != expectedId {
		return ErrIdMismatch
	}

	return nil
}

func (s *AuthService) ExtensionMiddleware() func(ctx *gin.Context) {

	type extToken struct {
		ChannelId string `json:"channel_id"`
		UserId    string `json:"user_id"`
		jwt.RegisteredClaims
	}

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

	type product struct {
		Rarity streampets.Rarity `json:"sku"`
	}

	type data struct {
		TransactionId uuid.UUID `json:"transactionId"`
		Product       product   `json:"product"`
	}

	type receipt struct {
		Data data `json:"data"`
		jwt.RegisteredClaims
	}

	return func(ctx *gin.Context) {
		request := new(request)
		if err := ctx.ShouldBindJSON(request); err != nil {
			slog.Error("failed to bind json: no 'receipt' field")
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}

		receipt := new(receipt)
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
