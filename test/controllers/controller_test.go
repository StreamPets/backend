package controllers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/controllers"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/services"
)

type CloseNotifierResponseWriter struct {
	*httptest.ResponseRecorder
}

func (c *CloseNotifierResponseWriter) CloseNotify() <-chan bool {
	return make(<-chan bool)
}

func TestHandleListen(t *testing.T) {
	setUpContext := func(channelID string, overlayID string) (*gin.Context, *CloseNotifierResponseWriter) {
		gin.SetMode(gin.TestMode)

		recorder := &CloseNotifierResponseWriter{httptest.NewRecorder()}
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest("GET", "/listen", nil)

		values := req.URL.Query()
		values.Add("channelID", channelID)
		values.Add("overlayID", overlayID)
		req.URL.RawQuery = values.Encode()

		ctx.Request = req
		return ctx, recorder
	}

	channelID := models.TwitchID("channel id")
	channelName := "channel name"
	overlayID := uuid.New()
	stream := make(services.EventStream)
	client := services.Client{
		ChannelName: channelName,
		Stream:      stream,
	}

	t.Run("handle listen functions correctly", func(t *testing.T) {
		ctx, recorder := setUpContext(string(channelID), overlayID.String())

		announcerMock := mock.Mock[services.Announcer]()
		authMock := mock.Mock[services.AuthService]()
		twitchMock := mock.Mock[repositories.TwitchRepository]()
		viewerMock := mock.Mock[services.DatabaseService]()

		mock.When(announcerMock.AddClient(channelName)).ThenReturn(client)
		mock.When(authMock.VerifyOverlayID(channelID, overlayID)).ThenReturn(nil)
		mock.When(twitchMock.GetUsername(channelID)).ThenReturn(channelName, nil)

		controller := controllers.NewController(
			announcerMock, authMock, twitchMock, viewerMock,
		)

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			controller.HandleListen(ctx)
		}()

		event := services.Event{Event: "event", Message: "message"}
		stream <- event
		close(stream)
		wg.Wait()

		mock.Verify(authMock, mock.Once()).VerifyOverlayID(channelID, overlayID)
		mock.Verify(announcerMock, mock.Once()).AddClient(channelName)
		mock.Verify(announcerMock, mock.Once()).RemoveClient(client)

		response := recorder.Body.String()
		if !strings.Contains(response, "event:event") {
			t.Errorf("expected event in response, got %s", response)
		}
		if !strings.Contains(response, "data:message") {
			t.Errorf("expected data in response, got %s", response)
		}
	})

	t.Run("handle listen functions correctly", func(t *testing.T) {
		ctx, recorder := setUpContext(string(channelID), overlayID.String())

		announcerMock := mock.Mock[services.Announcer]()
		authMock := mock.Mock[services.AuthService]()
		twitchMock := mock.Mock[repositories.TwitchRepository]()
		viewerMock := mock.Mock[services.DatabaseService]()

		mock.When(authMock.VerifyOverlayID(channelID, overlayID)).ThenReturn(services.ErrIdMismatch)

		controller := controllers.NewController(
			announcerMock, authMock, twitchMock, viewerMock,
		)

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			controller.HandleListen(ctx)
		}()

		wg.Wait()

		mock.Verify(authMock, mock.Once()).VerifyOverlayID(channelID, overlayID)

		response := recorder.Body.String()
		if !strings.Contains(response, services.ErrIdMismatch.Error()) {
			t.Errorf("expected event in response, got %s", response)
		}
	})
}

func TestGetStoreData(t *testing.T) {
	type Response struct {
		Owned []models.Item `json:"owned"`
		Store []models.Item `json:"store"`
	}

	setUpContext := func(tokenString string) (*gin.Context, *CloseNotifierResponseWriter) {
		gin.SetMode(gin.TestMode)

		// TODO: Test what happens without 'CloseNotifierResponseWriter'
		recorder := &CloseNotifierResponseWriter{httptest.NewRecorder()}
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest("GET", "/items", nil)

		req.Header.Add("x-extension-jwt", tokenString)

		ctx.Request = req
		return ctx, recorder
	}

	channelID := models.TwitchID("channel id")
	userID := models.TwitchID("user id")
	tokenString := "token string"
	token := services.ExtToken{ChannelID: channelID, UserID: userID}

	storeItems := []models.Item{{}}
	ownedItems := []models.Item{{}, {}}

	announcerMock := mock.Mock[services.Announcer]()
	authMock := mock.Mock[services.AuthService]()
	twitchMock := mock.Mock[repositories.TwitchRepository]()
	databaseMock := mock.Mock[services.DatabaseService]()

	mock.When(authMock.VerifyExtToken(tokenString)).ThenReturn(&token, nil)
	mock.When(databaseMock.GetTodaysItems(channelID)).ThenReturn(storeItems, nil)
	mock.When(databaseMock.GetOwnedItems(channelID, userID)).ThenReturn(ownedItems, nil)

	controller := controllers.NewController(
		announcerMock,
		authMock,
		twitchMock,
		databaseMock,
	)

	ctx, recorder := setUpContext(tokenString)
	controller.GetStoreData(ctx)

	mock.Verify(authMock, mock.Once()).VerifyExtToken(tokenString)
	mock.Verify(databaseMock, mock.Once()).GetTodaysItems(channelID)
	mock.Verify(databaseMock, mock.Once()).GetOwnedItems(channelID, userID)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected %d got %d", http.StatusOK, recorder.Code)
	}

	var response Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Error("could not parse json response")
	}

	assertSlicesEqual(t, ownedItems, response.Owned)
	assertSlicesEqual(t, storeItems, response.Store)
}

func assertSlicesEqual(t *testing.T, got, want []models.Item) {
	t.Helper()

	if !slices.Equal(got, want) {
		t.Errorf("expected %s got %s", want, got)
	}
}
