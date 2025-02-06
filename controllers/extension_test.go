package controllers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/services"
	"github.com/streampets/backend/twitch"
	"github.com/stretchr/testify/assert"
)

func TestSetSelectedItem(t *testing.T) {
	setUpContext := func(token, itemId string) *gin.Context {
		gin.SetMode(gin.TestMode)

		jsonData := []byte(fmt.Sprintf(`{
			"item_id": "%s"
		}`, itemId))

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		req, _ := http.NewRequest("POST", "/items", bytes.NewBuffer(jsonData))

		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
		req.Header.Add("x-extension-jwt", token)

		ctx.Request = req
		return ctx
	}

	t.Run("pet not updated when extension token is invalid", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.Id("channel id")
		tokenString := "token string"
		image := "image"

		userId := twitch.Id("user id")
		itemId := uuid.New()

		announcerMock := mock.Mock[UpdateAnnouncer]()
		verifierMock := mock.Mock[tokenVerifier]()
		storeMock := mock.Mock[StoreService]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(nil, assert.AnError)

		controller := NewExtensionController(
			announcerMock,
			verifierMock,
			storeMock,
		)

		controller.SetSelectedItem(setUpContext(tokenString, itemId.String()))

		mock.Verify(announcerMock, mock.Never()).AnnounceUpdate(channelId, userId, image)
	})

	t.Run("pet not updated when item id is not a valid uuid", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.Id("channel id")
		tokenString := "token string"
		image := "image"

		userId := twitch.Id("user id")
		itemId := "invalid id"

		announcerMock := mock.Mock[UpdateAnnouncer]()
		verifierMock := mock.Mock[tokenVerifier]()
		storeMock := mock.Mock[StoreService]()

		controller := NewExtensionController(
			announcerMock,
			verifierMock,
			storeMock,
		)

		controller.SetSelectedItem(setUpContext(tokenString, itemId))

		mock.Verify(announcerMock, mock.Never()).AnnounceUpdate(channelId, userId, image)
	})

	t.Run("pet not updated when item id does not exist", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.Id("channel id")
		tokenString := "token string"
		image := "image"

		userId := twitch.Id("user id")
		itemId := uuid.New()

		announcerMock := mock.Mock[UpdateAnnouncer]()
		verifierMock := mock.Mock[tokenVerifier]()
		storeMock := mock.Mock[StoreService]()

		mock.When(storeMock.GetItemById(itemId)).ThenReturn(nil, assert.AnError)

		controller := NewExtensionController(
			announcerMock,
			verifierMock,
			storeMock,
		)

		controller.SetSelectedItem(setUpContext(tokenString, itemId.String()))

		mock.Verify(announcerMock, mock.Never()).AnnounceUpdate(channelId, userId, image)
	})

	t.Run("pet not updated when item unowned", func(t *testing.T) {
		mock.SetUp(t)

		tokenString := "token string"
		image := "image"

		channelId := twitch.Id("channel id")
		userId := twitch.Id("user id")
		itemId := uuid.New()

		token := &services.ExtToken{
			ChannelId: channelId,
			UserId:    userId,
		}

		announcerMock := mock.Mock[UpdateAnnouncer]()
		verifierMock := mock.Mock[tokenVerifier]()
		storeMock := mock.Mock[StoreService]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(token, nil)
		mock.When(storeMock.SetSelectedItem(userId, channelId, itemId)).ThenReturn(assert.AnError)

		controller := NewExtensionController(
			announcerMock,
			verifierMock,
			storeMock,
		)

		controller.SetSelectedItem(setUpContext(tokenString, itemId.String()))

		mock.Verify(announcerMock, mock.Never()).AnnounceUpdate(channelId, userId, image)
	})

	t.Run("pet updated when pre-requisites are met", func(t *testing.T) {
		mock.SetUp(t)

		tokenString := "token string"
		image := "image"

		channelId := twitch.Id("channel id")
		userId := twitch.Id("user id")
		itemId := uuid.New()

		item := models.Item{ItemId: itemId, Image: image}

		token := services.ExtToken{
			ChannelId: channelId,
			UserId:    userId,
		}

		announcerMock := mock.Mock[UpdateAnnouncer]()
		verifierMock := mock.Mock[tokenVerifier]()
		storeMock := mock.Mock[StoreService]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(&token, nil)
		mock.When(storeMock.SetSelectedItem(userId, channelId, itemId)).ThenReturn(nil)
		mock.When(storeMock.GetItemById(itemId)).ThenReturn(item, nil)

		controller := NewExtensionController(
			announcerMock,
			verifierMock,
			storeMock,
		)

		controller.SetSelectedItem(setUpContext(tokenString, itemId.String()))

		mock.Verify(verifierMock, mock.Once()).VerifyExtToken(tokenString)
		mock.Verify(storeMock, mock.Once()).SetSelectedItem(userId, channelId, itemId)
		mock.Verify(announcerMock, mock.Once()).AnnounceUpdate(channelId, userId, image)
	})
}
