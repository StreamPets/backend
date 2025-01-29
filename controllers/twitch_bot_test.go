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
)

func TestAddUserToChannel(t *testing.T) {
	mock.SetUp(t)

	setUpContext := func(channelId, userId models.TwitchId, username string) *gin.Context {
		gin.SetMode(gin.TestMode)

		jsonData := []byte(fmt.Sprintf(`{
			"user_id": "%s",
			"username": "%s"
		}`, userId, username))

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		req, _ := http.NewRequest("", "", bytes.NewBuffer(jsonData))
		ctx.Params = gin.Params{{Key: ChannelId, Value: string(channelId)}}

		ctx.Request = req
		return ctx
	}

	channelId := models.TwitchId("channel id")
	userId := models.TwitchId("user id")
	username := "username"

	pet := services.Pet{Username: username}

	announcerMock := mock.Mock[Announcer]()
	itemsMock := mock.Mock[ItemGetSetter]()
	petsMock := mock.Mock[PetGetter]()

	mock.When(petsMock.GetPet(userId, channelId, username)).ThenReturn(pet, nil)

	controller := NewTwitchBotController(
		announcerMock,
		itemsMock,
		petsMock,
	)

	controller.AddPetToChannel(setUpContext(channelId, userId, username))

	mock.Verify(announcerMock, mock.Once()).AnnounceJoin(channelId, pet)
}

func TestRemoveUserFromChannel(t *testing.T) {
	mock.SetUp(t)

	setUpContext := func(channelId, userId models.TwitchId) *gin.Context {
		gin.SetMode(gin.TestMode)

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Params = gin.Params{
			{Key: ChannelId, Value: string(channelId)},
			{Key: UserId, Value: string(userId)},
		}

		return ctx
	}

	channelId := models.TwitchId("channel id")
	userId := models.TwitchId("user id")

	announcerMock := mock.Mock[Announcer]()
	itemsMock := mock.Mock[ItemGetSetter]()
	petsMock := mock.Mock[PetGetter]()

	controller := NewTwitchBotController(
		announcerMock,
		itemsMock,
		petsMock,
	)

	controller.RemoveUserFromChannel(setUpContext(channelId, userId))

	mock.Verify(announcerMock, mock.Once()).AnnouncePart(channelId, userId)
}

func TestAction(t *testing.T) {
	mock.SetUp(t)

	setUpContext := func(channelId, userId models.TwitchId, action string) *gin.Context {
		gin.SetMode(gin.TestMode)

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Params = gin.Params{
			{Key: ChannelId, Value: string(channelId)},
			{Key: UserId, Value: string(userId)},
			{Key: Action, Value: action},
		}

		return ctx
	}

	channelId := models.TwitchId("channel id")
	userId := models.TwitchId("user id")
	action := "action"

	announcerMock := mock.Mock[Announcer]()
	itemsMock := mock.Mock[ItemGetSetter]()
	petsMock := mock.Mock[PetGetter]()

	controller := NewTwitchBotController(
		announcerMock,
		itemsMock,
		petsMock,
	)

	controller.Action(setUpContext(channelId, userId, action))

	mock.Verify(announcerMock, mock.Once()).AnnounceAction(channelId, userId, action)
}

func TestUpdateUser(t *testing.T) {
	mock.SetUp(t)

	setUpContext := func(channelId, userId models.TwitchId, itemName string) *gin.Context {
		gin.SetMode(gin.TestMode)

		jsonData := []byte(fmt.Sprintf(`{
			"item_name": "%s"
		}`, itemName))

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		req, _ := http.NewRequest("", "", bytes.NewBuffer(jsonData))
		ctx.Params = gin.Params{
			{Key: ChannelId, Value: string(channelId)},
			{Key: UserId, Value: string(userId)},
		}
		ctx.Request = req

		return ctx
	}

	channelId := models.TwitchId("channel id")
	userId := models.TwitchId("user id")
	itemName := "item name"

	itemId := uuid.New()
	image := "image"

	item := models.Item{ItemId: itemId, Image: image}

	announcerMock := mock.Mock[Announcer]()
	itemsMock := mock.Mock[ItemGetSetter]()
	petsMock := mock.Mock[PetGetter]()

	mock.When(itemsMock.GetItemByName(channelId, itemName)).ThenReturn(item, nil)

	controller := NewTwitchBotController(
		announcerMock,
		itemsMock,
		petsMock,
	)

	controller.UpdateUser(setUpContext(channelId, userId, itemName))

	mock.Verify(itemsMock, mock.Once()).SetSelectedItem(userId, channelId, itemId)
	mock.Verify(announcerMock, mock.Once()).AnnounceUpdate(channelId, userId, image)
}
