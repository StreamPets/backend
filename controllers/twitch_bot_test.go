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

	setUpContext := func(userId models.TwitchId, channelName, username string) *gin.Context {
		gin.SetMode(gin.TestMode)

		jsonData := []byte(fmt.Sprintf(`{
			"user_id": "%s",
			"username": "%s"
		}`, userId, username))

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		req, _ := http.NewRequest("", "", bytes.NewBuffer(jsonData))
		ctx.Params = gin.Params{{Key: "channelName", Value: channelName}}

		ctx.Request = req
		return ctx
	}

	userId := models.TwitchId("user id")
	channelId := models.TwitchId("channel id")
	username := "username"
	channelName := "channel name"

	pet := services.Pet{Username: username}

	announcerMock := mock.Mock[Announcer]()
	itemsMock := mock.Mock[ItemGetSetter]()
	petsMock := mock.Mock[PetGetter]()
	usersMock := mock.Mock[UserIdGetter]()

	mock.When(usersMock.GetUserId(channelName)).ThenReturn(channelId, nil)
	mock.When(petsMock.GetPet(userId, channelId, username)).ThenReturn(pet, nil)

	controller := NewTwitchBotController(
		announcerMock,
		itemsMock,
		petsMock,
		usersMock,
	)

	controller.AddPetToChannel(setUpContext(userId, channelName, username))

	mock.Verify(announcerMock, mock.Once()).AnnounceJoin(channelName, pet)
}

func TestRemoveUserFromChannel(t *testing.T) {
	mock.SetUp(t)

	setUpContext := func(userId models.TwitchId, channelName string) *gin.Context {
		gin.SetMode(gin.TestMode)

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Params = gin.Params{
			{Key: ChannelName, Value: channelName},
			{Key: UserId, Value: string(userId)},
		}

		return ctx
	}

	userId := models.TwitchId("user id")
	channelId := models.TwitchId("channel id")
	channelName := "channel name"

	announcerMock := mock.Mock[Announcer]()
	itemsMock := mock.Mock[ItemGetSetter]()
	petsMock := mock.Mock[PetGetter]()
	usersMock := mock.Mock[UserIdGetter]()

	mock.When(usersMock.GetUserId(channelName)).ThenReturn(channelId, nil)

	controller := NewTwitchBotController(
		announcerMock,
		itemsMock,
		petsMock,
		usersMock,
	)

	controller.RemoveUserFromChannel(setUpContext(userId, channelName))

	mock.Verify(announcerMock, mock.Once()).AnnouncePart(channelName, userId)
}

func TestAction(t *testing.T) {
	mock.SetUp(t)

	setUpContext := func(userId models.TwitchId, channelName, action string) *gin.Context {
		gin.SetMode(gin.TestMode)

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Params = gin.Params{
			{Key: ChannelName, Value: channelName},
			{Key: UserId, Value: string(userId)},
			{Key: Action, Value: action},
		}

		return ctx
	}

	userId := models.TwitchId("user id")
	channelName := "channel name"
	action := "action"

	announcerMock := mock.Mock[Announcer]()
	itemsMock := mock.Mock[ItemGetSetter]()
	petsMock := mock.Mock[PetGetter]()
	usersMock := mock.Mock[UserIdGetter]()

	controller := NewTwitchBotController(
		announcerMock,
		itemsMock,
		petsMock,
		usersMock,
	)

	controller.Action(setUpContext(userId, channelName, action))

	mock.Verify(announcerMock, mock.Once()).AnnounceAction(channelName, action, userId)
}

func TestUpdateUser(t *testing.T) {
	mock.SetUp(t)

	setUpContext := func(userId models.TwitchId, channelName, itemName string) *gin.Context {
		gin.SetMode(gin.TestMode)

		jsonData := []byte(fmt.Sprintf(`{
			"item_name": "%s"
		}`, itemName))

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		req, _ := http.NewRequest("", "", bytes.NewBuffer(jsonData))
		ctx.Params = gin.Params{
			{Key: ChannelName, Value: channelName},
			{Key: UserId, Value: string(userId)},
		}
		ctx.Request = req

		return ctx
	}

	userId := models.TwitchId("user id")
	channelId := models.TwitchId("channel id")
	channelName := "channel name"
	itemName := "item name"

	itemId := uuid.New()
	image := "image"

	item := models.Item{ItemId: itemId, Image: image}

	announcerMock := mock.Mock[Announcer]()
	itemsMock := mock.Mock[ItemGetSetter]()
	petsMock := mock.Mock[PetGetter]()
	usersMock := mock.Mock[UserIdGetter]()

	mock.When(usersMock.GetUserId(channelName)).ThenReturn(channelId, nil)
	mock.When(itemsMock.GetItemByName(channelId, itemName)).ThenReturn(item, nil)

	controller := NewTwitchBotController(
		announcerMock,
		itemsMock,
		petsMock,
		usersMock,
	)

	controller.UpdateUser(setUpContext(userId, channelName, itemName))

	mock.Verify(itemsMock, mock.Once()).SetSelectedItem(userId, channelId, itemId)
	mock.Verify(announcerMock, mock.Once()).AnnounceUpdate(channelName, image, userId)
}
