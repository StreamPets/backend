package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/announcers"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/services"
	"github.com/streampets/backend/test"
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

		overlays := mock.Mock[overlayIdGetter]()
		validator := mock.Mock[tokenValidator]()

		ctx, recorder := setUpContext("")
		handleLogin(validator, overlays)(ctx)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run("unauthorized status when access token invalid", func(t *testing.T) {
		mock.SetUp(t)

		invalidToken := "inavlid token"
		ctx, recorder := setUpContext(invalidToken)

		overlays := mock.Mock[overlayIdGetter]()
		validator := mock.Mock[tokenValidator]()

		mock.When(validator.ValidateToken(ctx, invalidToken)).ThenReturn(nil, twitch.ErrInvalidUserToken)

		handleLogin(validator, overlays)(ctx)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run("internal server error when validate token fails", func(t *testing.T) {
		mock.SetUp(t)

		invalidToken := "inavlid token"
		ctx, recorder := setUpContext(invalidToken)

		overlays := mock.Mock[overlayIdGetter]()
		validator := mock.Mock[tokenValidator]()

		mock.When(validator.ValidateToken(ctx, invalidToken)).ThenReturn(nil, assert.AnError)

		handleLogin(validator, overlays)(ctx)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("status bad request when channel id has no overlay id", func(t *testing.T) {
		mock.SetUp(t)

		token := "token"
		channelId := twitch.Id("channel id")
		ctx, recorder := setUpContext(token)

		overlays := mock.Mock[overlayIdGetter]()
		validator := mock.Mock[tokenValidator]()

		mock.When(validator.ValidateToken(ctx, token)).ThenReturn(channelId, nil)
		mock.When(overlays.GetOverlayId(channelId)).ThenReturn(nil, repositories.NewErrNoOverlayId(channelId))

		handleLogin(validator, overlays)(ctx)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("status bad request when channel id has no overlay id", func(t *testing.T) {
		mock.SetUp(t)

		token := "token"
		channelId := twitch.Id("channel id")
		ctx, recorder := setUpContext(token)

		overlays := mock.Mock[overlayIdGetter]()
		validator := mock.Mock[tokenValidator]()

		mock.When(validator.ValidateToken(ctx, token)).ThenReturn(channelId, nil)
		mock.When(overlays.GetOverlayId(channelId)).ThenReturn(nil, assert.AnError)

		handleLogin(validator, overlays)(ctx)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("channel id and overlay id returned in normal case", func(t *testing.T) {
		mock.SetUp(t)

		type userData struct {
			OverlayId uuid.UUID `json:"overlay_id"`
			ChannelId twitch.Id `json:"channel_id"`
		}

		token := "token"
		channelId := twitch.Id("channel id")
		overlayId := uuid.New()
		ctx, recorder := setUpContext(token)

		overlays := mock.Mock[overlayIdGetter]()
		validator := mock.Mock[tokenValidator]()

		mock.When(validator.ValidateToken(ctx, token)).ThenReturn(channelId, nil)
		mock.When(overlays.GetOverlayId(channelId)).ThenReturn(overlayId, nil)

		handleLogin(validator, overlays)(ctx)

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

func TestHandleListen(t *testing.T) {
	setUpContext := func(channelId twitch.Id, overlayId uuid.UUID) (*gin.Context, *test.CloseNotifierResponseWriter) {
		gin.SetMode(gin.TestMode)

		recorder := &test.CloseNotifierResponseWriter{ResponseRecorder: httptest.NewRecorder()}
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest("GET", "/listen", nil)

		values := req.URL.Query()
		values.Add("channelId", string(channelId))
		values.Add("overlayId", overlayId.String())
		req.URL.RawQuery = values.Encode()

		ctx.Request = req
		return ctx, recorder
	}

	channelId := twitch.Id("channel id")
	overlayId := uuid.New()

	t.Run("receive and send events from stream", func(t *testing.T) {
		mock.SetUp(t)

		ctx, recorder := setUpContext(channelId, overlayId)

		stream := make(chan announcers.Announcement)
		client := announcers.Client{Stream: stream}

		announcerMock := mock.Mock[clientAddRemover]()
		authMock := mock.Mock[overlayIdValidator]()

		mock.When(announcerMock.AddClient(channelId)).ThenReturn(client)

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			handleListen(announcerMock, authMock)(ctx)
		}()

		stream <- announcers.Announcement{
			Event:   "event",
			Message: "message",
		}

		close(stream)
		wg.Wait()

		mock.Verify(authMock, mock.Once()).ValidateOverlayId(channelId, overlayId)
		mock.Verify(announcerMock, mock.Once()).AddClient(channelId)
		mock.Verify(announcerMock, mock.Once()).RemoveClient(client)

		assert.Contains(t, recorder.Body.String(), "event:event")
		assert.Contains(t, recorder.Body.String(), "data:message")
	})

	t.Run("client not added when overlay id and channel id do not match", func(t *testing.T) {
		mock.SetUp(t)

		ctx, recorder := setUpContext(channelId, overlayId)

		announcerMock := mock.Mock[clientAddRemover]()
		authMock := mock.Mock[overlayIdValidator]()

		mock.When(authMock.ValidateOverlayId(channelId, overlayId)).ThenReturn(services.ErrIdMismatch)

		handleListen(announcerMock, authMock)(ctx)

		mock.Verify(authMock, mock.Once()).ValidateOverlayId(channelId, overlayId)
		mock.Verify(announcerMock, mock.Never()).AddClient(channelId)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})
}

func TestGetStoreData(t *testing.T) {
	setUpContext := func(tokenString string) (*gin.Context, *httptest.ResponseRecorder) {
		gin.SetMode(gin.TestMode)

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest("GET", "/items", nil)

		req.Header.Add(XExtensionJwt, tokenString)

		ctx.Request = req
		return ctx, recorder
	}

	t.Run("access forbidden with invalid extension token", func(t *testing.T) {
		mock.SetUp(t)

		tokenString := "invalid token"

		verifierMock := mock.Mock[extTokenVerifier]()
		storeMock := mock.Mock[channelItemGetter]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(nil, services.ErrInvalidToken)

		ctx, recorder := setUpContext(tokenString)
		handleGetStoreData(verifierMock, storeMock)(ctx)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run("internal server error when error received from verify token", func(t *testing.T) {
		mock.SetUp(t)

		tokenString := "invalid token"

		verifierMock := mock.Mock[extTokenVerifier]()
		storeMock := mock.Mock[channelItemGetter]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(nil, assert.AnError)

		ctx, recorder := setUpContext(tokenString)
		handleGetStoreData(verifierMock, storeMock)(ctx)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("internal server error when error received from get channels items", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.Id("channel id")
		userId := twitch.Id("user id")
		tokenString := "token string"
		token := services.ExtToken{ChannelId: channelId, UserId: userId}

		storeItems := []models.Item{{}, {}}

		verifierMock := mock.Mock[extTokenVerifier]()
		storeMock := mock.Mock[channelItemGetter]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(&token, nil)
		mock.When(storeMock.GetChannelsItems(channelId)).ThenReturn(storeItems, assert.AnError)

		ctx, recorder := setUpContext(tokenString)
		handleGetStoreData(verifierMock, storeMock)(ctx)

		mock.Verify(verifierMock, mock.Once()).VerifyExtToken(tokenString)
		mock.Verify(storeMock, mock.Once()).GetChannelsItems(channelId)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("items returned when extension token and channel id are valid", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.Id("channel id")
		userId := twitch.Id("user id")
		tokenString := "token string"
		token := services.ExtToken{ChannelId: channelId, UserId: userId}

		storeItems := []models.Item{{}, {}}

		verifierMock := mock.Mock[extTokenVerifier]()
		storeMock := mock.Mock[channelItemGetter]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(&token, nil)
		mock.When(storeMock.GetChannelsItems(channelId)).ThenReturn(storeItems, nil)

		ctx, recorder := setUpContext(tokenString)
		handleGetStoreData(verifierMock, storeMock)(ctx)

		mock.Verify(verifierMock, mock.Once()).VerifyExtToken(tokenString)
		mock.Verify(storeMock, mock.Once()).GetChannelsItems(channelId)

		assert.Equal(t, recorder.Code, http.StatusOK)

		var response []models.Item
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Errorf("could not parse json response")
		}

		assert.Equal(t, storeItems, response)
	})
}
