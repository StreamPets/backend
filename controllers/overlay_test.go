package controllers

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/services"
	"github.com/stretchr/testify/assert"
)

type CloseNotifierResponseWriter struct {
	*httptest.ResponseRecorder
}

func (c *CloseNotifierResponseWriter) CloseNotify() <-chan bool {
	return make(<-chan bool)
}

func TestHandleListen(t *testing.T) {
	setUpContext := func(channelId, overlayId string) (*gin.Context, *CloseNotifierResponseWriter) {
		gin.SetMode(gin.TestMode)

		recorder := &CloseNotifierResponseWriter{httptest.NewRecorder()}
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest("GET", "/listen", nil)

		values := req.URL.Query()
		values.Add("channelId", channelId)
		values.Add("overlayId", overlayId)
		req.URL.RawQuery = values.Encode()

		ctx.Request = req
		return ctx, recorder
	}

	channelId := models.TwitchId("channel id")
	overlayId := uuid.New()
	channelName := "channel name"

	t.Run("receive and send events from stream", func(t *testing.T) {
		mock.SetUp(t)

		ctx, recorder := setUpContext(string(channelId), overlayId.String())

		stream := make(services.EventStream)
		client := services.Client{
			ChannelName: channelName,
			Stream:      stream,
		}

		clientsMock := mock.Mock[ClientAddRemover]()
		verifierMock := mock.Mock[OverlayIdVerifier]()
		userMock := mock.Mock[UsernameGetter]()

		mock.When(clientsMock.AddClient(channelName)).ThenReturn(client)
		mock.When(userMock.GetUsername(channelId)).ThenReturn(channelName, nil)

		controller := NewOverlayController(
			clientsMock,
			verifierMock,
			userMock,
		)

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			controller.HandleListen(ctx)
		}()

		stream <- services.Event{Event: "event", Message: "message"}

		close(stream)
		wg.Wait()

		mock.Verify(verifierMock, mock.Once()).VerifyOverlayId(channelId, overlayId)
		mock.Verify(clientsMock, mock.Once()).AddClient(channelName)
		mock.Verify(clientsMock, mock.Once()).RemoveClient(client)

		assert.Contains(t, recorder.Body.String(), "event:event")
		assert.Contains(t, recorder.Body.String(), "data:message")
	})

	t.Run("client not added when overlay id and channel id do not match", func(t *testing.T) {
		mock.SetUp(t)

		ctx, recorder := setUpContext(string(channelId), overlayId.String())

		clientsMock := mock.Mock[ClientAddRemover]()
		verifierMock := mock.Mock[OverlayIdVerifier]()
		userMock := mock.Mock[UsernameGetter]()

		mock.When(verifierMock.VerifyOverlayId(channelId, overlayId)).ThenReturn(services.ErrIdMismatch)

		controller := NewOverlayController(
			clientsMock,
			verifierMock,
			userMock,
		)

		controller.HandleListen(ctx)

		mock.Verify(verifierMock, mock.Once()).VerifyOverlayId(channelId, overlayId)
		mock.Verify(clientsMock, mock.Never()).AddClient(channelName)

		assert.Contains(t, recorder.Body.String(), services.ErrIdMismatch.Error())
	})
}
