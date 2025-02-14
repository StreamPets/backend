package auth

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/twitch"
	"github.com/stretchr/testify/assert"
)

func TestVerifyOverlayId(t *testing.T) {

	type MockDep interface {
		GetOverlayId(channelId twitch.Id) (uuid.UUID, error)
	}

	t.Run("status ok when overlay and channel ids match", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.Id("channel id")
		overlayId := uuid.New()

		mockDep := mock.Mock[MockDep]()
		mock.When(mockDep.GetOverlayId(channelId)).ThenReturn(overlayId, nil)

		router := gin.Default()
		router.Use(ListenAuthentication(mockDep.GetOverlayId))
		router.GET("/protected", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, nil)
		})

		req := httptest.NewRequest("GET", "/protected", nil)

		values := req.URL.Query()
		values.Add(ChannelId, string(channelId))
		values.Add(OverlayId, overlayId.String())
		req.URL.RawQuery = values.Encode()

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		mock.Verify(mockDep, mock.Once()).GetOverlayId(channelId)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("unauthorized when overlay and channel ids do not match", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.Id("channel id")
		overlayId := uuid.New()

		mockDep := mock.Mock[MockDep]()
		mock.When(mockDep.GetOverlayId(channelId)).ThenReturn(nil, ErrIdMismatch)

		router := gin.Default()
		router.Use(ListenAuthentication(mockDep.GetOverlayId))
		router.GET("/protected", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, nil)
		})

		req := httptest.NewRequest("GET", "/protected", nil)

		values := req.URL.Query()
		values.Add(ChannelId, string(channelId))
		values.Add(OverlayId, overlayId.String())
		req.URL.RawQuery = values.Encode()

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		mock.Verify(mockDep, mock.Once()).GetOverlayId(channelId)
		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})
}

func TestExtensionMiddleware(t *testing.T) {

	clientSecret := "secret"
	channelId := twitch.Id("channel id")
	userId := twitch.Id("user id")

	authService := New(clientSecret)

	router := gin.Default()
	router.Use(authService.ExtensionMiddleware())
	router.GET("/protected", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, nil)
	})

	t.Run("valid", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"channel_id": channelId,
			"user_id":    userId,
		})

		tokenString, err := token.SignedString([]byte(clientSecret))
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Add(XExtensionJwt, tokenString)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("invalid", func(t *testing.T) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"channel_id": channelId,
			"user_id":    userId,
			"exp":        0,
		})

		tokenString, err := token.SignedString([]byte(clientSecret))
		assert.NoError(t, err)

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Add(XExtensionJwt, tokenString)
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})
}

func TestReceiptMiddleware(t *testing.T) {

	clientSecret := "secret"
	authService := New(clientSecret)

	router := gin.Default()
	router.Use(authService.ReceiptMiddleware())
	router.GET("/protected", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, nil)
	})

	t.Run("valid", func(t *testing.T) {
		mock.SetUp(t)

		transactionId := uuid.New()

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"data": map[string]interface{}{
				"transactionId": transactionId,
				"product": map[string]interface{}{
					"sku": "common",
				},
			},
		})

		receiptString, err := token.SignedString([]byte(clientSecret))
		assert.NoError(t, err)

		jsonData := []byte(fmt.Sprintf(`{
			"receipt": "%s"
		}`, receiptString))

		req := httptest.NewRequest("GET", "/protected", bytes.NewBuffer(jsonData))

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("expired", func(t *testing.T) {
		mock.SetUp(t)

		transactionId := uuid.New()

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"data": map[string]interface{}{
				"transactionId": transactionId,
				"product": map[string]interface{}{
					"sku": "common",
				},
			},
			"exp": 0,
		})

		receiptString, err := token.SignedString([]byte(clientSecret))
		assert.NoError(t, err)

		jsonData := []byte(fmt.Sprintf(`{
			"receipt": "%s"
		}`, receiptString))

		req := httptest.NewRequest("GET", "/protected", bytes.NewBuffer(jsonData))

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})
}
