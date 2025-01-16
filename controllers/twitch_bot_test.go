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

func TestAddViewerToChannel(t *testing.T) {
	mock.SetUp(t)

	setUpContext := func(viewerId models.UserId, channelName, username string) *gin.Context {
		gin.SetMode(gin.TestMode)

		jsonData := []byte(fmt.Sprintf(`{
			"viewer_id": "%s",
			"username": "%s"
		}`, viewerId, username))

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		req, _ := http.NewRequest("", "", bytes.NewBuffer(jsonData))
		ctx.Params = gin.Params{{Key: "channelName", Value: channelName}}

		ctx.Request = req
		return ctx
	}

	viewerId := models.UserId("viewer id")
	channelId := models.UserId("channel id")
	username := "username"
	channelName := "channel name"

	viewer := services.Pet{Username: username}

	announcerMock := mock.Mock[Announcer]()
	itemsMock := mock.Mock[ItemGetSetter]()
	petsMock := mock.Mock[PetGetter]()
	usersMock := mock.Mock[UserIdGetter]()

	mock.When(usersMock.GetUserId(channelName)).ThenReturn(channelId, nil)
	mock.When(petsMock.GetPet(viewerId, channelId, username)).ThenReturn(viewer, nil)

	controller := NewTwitchBotController(
		announcerMock,
		itemsMock,
		petsMock,
		usersMock,
	)

	controller.AddViewerToChannel(setUpContext(viewerId, channelName, username))

	mock.Verify(announcerMock, mock.Once()).AnnounceJoin(channelName, viewer)
}

func TestRemoveViewerFromChannel(t *testing.T) {
	mock.SetUp(t)

	setUpContext := func(viewerId models.UserId, channelName string) *gin.Context {
		gin.SetMode(gin.TestMode)

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Params = gin.Params{
			{Key: ChannelName, Value: channelName},
			{Key: ViewerId, Value: string(viewerId)},
		}

		return ctx
	}

	viewerId := models.UserId("viewer id")
	channelId := models.UserId("channel id")
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

	controller.RemoveViewerFromChannel(setUpContext(viewerId, channelName))

	mock.Verify(announcerMock, mock.Once()).AnnouncePart(channelName, viewerId)
}

func TestAction(t *testing.T) {
	mock.SetUp(t)

	setUpContext := func(viewerId models.UserId, channelName, action string) *gin.Context {
		gin.SetMode(gin.TestMode)

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Params = gin.Params{
			{Key: ChannelName, Value: channelName},
			{Key: ViewerId, Value: string(viewerId)},
			{Key: Action, Value: action},
		}

		return ctx
	}

	viewerId := models.UserId("viewer id")
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

	controller.Action(setUpContext(viewerId, channelName, action))

	mock.Verify(announcerMock, mock.Once()).AnnounceAction(channelName, action, viewerId)
}

func TestUpdateViewer(t *testing.T) {
	mock.SetUp(t)

	setUpContext := func(viewerId models.UserId, channelName, itemName string) *gin.Context {
		gin.SetMode(gin.TestMode)

		jsonData := []byte(fmt.Sprintf(`{
			"item_name": "%s"
		}`, itemName))

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		req, _ := http.NewRequest("", "", bytes.NewBuffer(jsonData))
		ctx.Params = gin.Params{
			{Key: ChannelName, Value: channelName},
			{Key: ViewerId, Value: string(viewerId)},
		}
		ctx.Request = req

		return ctx
	}

	viewerId := models.UserId("viewer id")
	channelId := models.UserId("channel id")
	channelName := "channel name"
	itemName := "item name"

	itemId := uuid.New()
	image := "image"

	item := models.Item{ItemId: itemId, Image: image}

	announcerMock := mock.Mock[Announcer]()
	itemsMock := mock.Mock[ItemGetSetter]()
	viewersMock := mock.Mock[PetGetter]()
	usersMock := mock.Mock[UserIdGetter]()

	mock.When(usersMock.GetUserId(channelName)).ThenReturn(channelId, nil)
	mock.When(itemsMock.GetItemByName(channelId, itemName)).ThenReturn(item, nil)

	controller := NewTwitchBotController(
		announcerMock,
		itemsMock,
		viewersMock,
		usersMock,
	)

	controller.UpdateViewer(setUpContext(viewerId, channelName, itemName))

	mock.Verify(itemsMock, mock.Once()).SetSelectedItem(viewerId, channelId, itemId)
	mock.Verify(announcerMock, mock.Once()).AnnounceUpdate(channelName, image, viewerId)
}
