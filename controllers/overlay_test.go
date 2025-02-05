package controllers

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/announcers"
	"github.com/streampets/backend/services"
	"github.com/streampets/backend/twitch"
	"github.com/stretchr/testify/assert"
)

type CloseNotifierResponseWriter struct {
	*httptest.ResponseRecorder
}

func (c *CloseNotifierResponseWriter) CloseNotify() <-chan bool {
	return make(<-chan bool)
}

func TestHandleListen(t *testing.T) {
	setUpContext := func(channelId twitch.Id, overlayId uuid.UUID) (*gin.Context, *CloseNotifierResponseWriter) {
		gin.SetMode(gin.TestMode)

		recorder := &CloseNotifierResponseWriter{httptest.NewRecorder()}
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
		verifierMock := mock.Mock[OverlayIdVerifier]()

		mock.When(announcerMock.AddClient(channelId)).ThenReturn(client)

		controller := NewOverlayController(
			announcerMock,
			verifierMock,
		)

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			controller.HandleListen(ctx)
		}()

		stream <- announcers.Announcement{
			Event:   "event",
			Message: "message",
		}

		close(stream)
		wg.Wait()

		mock.Verify(verifierMock, mock.Once()).VerifyOverlayId(channelId, overlayId)
		mock.Verify(announcerMock, mock.Once()).AddClient(channelId)
		mock.Verify(announcerMock, mock.Once()).RemoveClient(client)

		assert.Contains(t, recorder.Body.String(), "event:event")
		assert.Contains(t, recorder.Body.String(), "data:message")
	})

	t.Run("client not added when overlay id and channel id do not match", func(t *testing.T) {
		mock.SetUp(t)

		ctx, recorder := setUpContext(channelId, overlayId)

		clientMock := mock.Mock[clientAddRemover]()
		verifierMock := mock.Mock[OverlayIdVerifier]()

		mock.When(verifierMock.VerifyOverlayId(channelId, overlayId)).ThenReturn(services.ErrIdMismatch)

		controller := NewOverlayController(
			clientMock,
			verifierMock,
		)

		controller.HandleListen(ctx)

		mock.Verify(verifierMock, mock.Once()).VerifyOverlayId(channelId, overlayId)
		mock.Verify(clientMock, mock.Never()).AddClient(channelId)

		assert.Contains(t, recorder.Body.String(), services.ErrIdMismatch.Error())
	})
}
