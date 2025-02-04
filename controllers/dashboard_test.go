package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/twitch"
	"github.com/stretchr/testify/assert"
)

func TestHandleLogin(t *testing.T) {
	setUpContext := func(cookie string) (*gin.Context, *httptest.ResponseRecorder) {
		gin.SetMode(gin.TestMode)

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest("GET", "/items", nil)

		if cookie != "" {
			req.AddCookie(&http.Cookie{
				Name:  "Authorization",
				Value: cookie,
			})
		}

		ctx.Request = req
		return ctx, recorder
	}

	t.Run("unauthorized status when no 'Authorization' cookie present", func(t *testing.T) {
		mock.SetUp(t)

		overlays := mock.Mock[OverlayIdGetter]()
		validator := mock.Mock[TokenValidator]()

		controller := NewDashboardController(overlays, validator)

		ctx, recorder := setUpContext("")
		controller.HandleLogin(ctx)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run("unauthorized status when access token invalid", func(t *testing.T) {
		mock.SetUp(t)

		invalidToken := "inavlid token"

		overlays := mock.Mock[OverlayIdGetter]()
		validator := mock.Mock[TokenValidator]()

		mock.When(validator.ValidateToken(invalidToken)).ThenReturn(nil, twitch.ErrInvalidAccessToken)

		controller := NewDashboardController(overlays, validator)

		ctx, recorder := setUpContext(invalidToken)
		controller.HandleLogin(ctx)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run("internal server error when validate token fails", func(t *testing.T) {
		mock.SetUp(t)

		invalidToken := "inavlid token"

		overlays := mock.Mock[OverlayIdGetter]()
		validator := mock.Mock[TokenValidator]()

		mock.When(validator.ValidateToken(invalidToken)).ThenReturn(nil, assert.AnError)

		controller := NewDashboardController(overlays, validator)

		ctx, recorder := setUpContext(invalidToken)
		controller.HandleLogin(ctx)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("status bad request when channel id has no overlay id", func(t *testing.T) {
		mock.SetUp(t)

		token := "token"
		channelId := models.TwitchId("channel id")
		tokenResponse := twitch.ValidateTokenResponse{UserId: channelId}

		overlays := mock.Mock[OverlayIdGetter]()
		validator := mock.Mock[TokenValidator]()

		mock.When(validator.ValidateToken(token)).ThenReturn(tokenResponse, nil)
		mock.When(overlays.GetOverlayId(channelId)).ThenReturn(nil, repositories.NewErrNoOverlayId(channelId))

		controller := NewDashboardController(overlays, validator)

		ctx, recorder := setUpContext(token)
		controller.HandleLogin(ctx)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("status bad request when channel id has no overlay id", func(t *testing.T) {
		mock.SetUp(t)

		token := "token"
		channelId := models.TwitchId("channel id")
		tokenResponse := twitch.ValidateTokenResponse{UserId: channelId}

		overlays := mock.Mock[OverlayIdGetter]()
		validator := mock.Mock[TokenValidator]()

		mock.When(validator.ValidateToken(token)).ThenReturn(tokenResponse, nil)
		mock.When(overlays.GetOverlayId(channelId)).ThenReturn(nil, assert.AnError)

		controller := NewDashboardController(overlays, validator)

		ctx, recorder := setUpContext(token)
		controller.HandleLogin(ctx)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("channel id and overlay id returned in normal case", func(t *testing.T) {
		mock.SetUp(t)

		type userData struct {
			OverlayId uuid.UUID       `json:"overlay_id"`
			ChannelId models.TwitchId `json:"channel_id"`
		}

		token := "token"
		channelId := models.TwitchId("channel id")
		overlayId := uuid.New()
		tokenResponse := twitch.ValidateTokenResponse{UserId: channelId}

		overlays := mock.Mock[OverlayIdGetter]()
		validator := mock.Mock[TokenValidator]()

		mock.When(validator.ValidateToken(token)).ThenReturn(tokenResponse, nil)
		mock.When(overlays.GetOverlayId(channelId)).ThenReturn(overlayId, nil)

		controller := NewDashboardController(overlays, validator)

		ctx, recorder := setUpContext(token)
		controller.HandleLogin(ctx)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var actual userData
		if err := json.Unmarshal(recorder.Body.Bytes(), &actual); err != nil {
			t.Errorf("could not parse json response")
		}

		expected := userData{
			OverlayId: overlayId,
			ChannelId: channelId,
		}

		assert.Equal(t, expected, actual)
	})
}
