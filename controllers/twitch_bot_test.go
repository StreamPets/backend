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

	setUpContext := func(userID models.TwitchID, channelName, username string) *gin.Context {
		gin.SetMode(gin.TestMode)

		jsonData := []byte(fmt.Sprintf(`{
			"user_id": "%s",
			"username": "%s"
		}`, userID, username))

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		req, _ := http.NewRequest("", "", bytes.NewBuffer(jsonData))
		ctx.Params = gin.Params{{Key: "channelName", Value: channelName}}

		ctx.Request = req
		return ctx
	}

	userID := models.TwitchID("user id")
	channelID := models.TwitchID("channel id")
	username := "username"
	channelName := "channel name"

	viewer := services.Viewer{Username: username}

	announcerMock := mock.Mock[Announcer]()
	itemsMock := mock.Mock[ItemGetSetter]()
	viewersMock := mock.Mock[ViewerGetter]()
	usersMock := mock.Mock[UserIDGetter]()
	cacheMock := mock.Mock[ViewerCache]()

	mock.When(usersMock.GetUserID(channelName)).ThenReturn(channelID, nil)
	mock.When(viewersMock.GetViewer(userID, channelID, username)).ThenReturn(viewer, nil)

	controller := NewTwitchBotController(
		announcerMock,
		itemsMock,
		viewersMock,
		usersMock,
		cacheMock,
	)

	controller.AddViewerToChannel(setUpContext(userID, channelName, username))

	mock.Verify(announcerMock, mock.Once()).AnnounceJoin(channelName, viewer)
	mock.Verify(cacheMock, mock.Once()).AddViewer(channelID, viewer)
}

func TestRemoveViewerFromChannel(t *testing.T) {
	mock.SetUp(t)

	setUpContext := func(userID models.TwitchID, channelName string) *gin.Context {
		gin.SetMode(gin.TestMode)

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Params = gin.Params{
			{Key: "channelName", Value: channelName},
			{Key: "userID", Value: string(userID)},
		}

		return ctx
	}

	userID := models.TwitchID("user id")
	channelID := models.TwitchID("channel ID")
	channelName := "channel name"

	announcerMock := mock.Mock[Announcer]()
	itemsMock := mock.Mock[ItemGetSetter]()
	viewersMock := mock.Mock[ViewerGetter]()
	usersMock := mock.Mock[UserIDGetter]()
	cacheMock := mock.Mock[ViewerCache]()

	mock.When(usersMock.GetUserID(channelName)).ThenReturn(channelID, nil)

	controller := NewTwitchBotController(
		announcerMock,
		itemsMock,
		viewersMock,
		usersMock,
		cacheMock,
	)

	controller.RemoveViewerFromChannel(setUpContext(userID, channelName))

	mock.Verify(announcerMock, mock.Once()).AnnouncePart(channelName, userID)
	mock.Verify(cacheMock, mock.Once()).RemoveViewer(channelID, userID)
}

func TestAction(t *testing.T) {
	mock.SetUp(t)

	setUpContext := func(userID models.TwitchID, channelName, action string) *gin.Context {
		gin.SetMode(gin.TestMode)

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Params = gin.Params{
			{Key: "channelName", Value: channelName},
			{Key: "userID", Value: string(userID)},
			{Key: "action", Value: action},
		}

		return ctx
	}

	userID := models.TwitchID("user id")
	channelName := "channel name"
	action := "action"

	announcerMock := mock.Mock[Announcer]()
	itemsMock := mock.Mock[ItemGetSetter]()
	viewersMock := mock.Mock[ViewerGetter]()
	usersMock := mock.Mock[UserIDGetter]()
	cacheMock := mock.Mock[ViewerCache]()

	controller := NewTwitchBotController(
		announcerMock,
		itemsMock,
		viewersMock,
		usersMock,
		cacheMock,
	)

	controller.Action(setUpContext(userID, channelName, action))

	mock.Verify(announcerMock, mock.Once()).AnnounceAction(channelName, action, userID)
}

func TestUpdateViewer(t *testing.T) {
	mock.SetUp(t)

	setUpContext := func(userID models.TwitchID, channelName, itemName string) *gin.Context {
		gin.SetMode(gin.TestMode)

		jsonData := []byte(fmt.Sprintf(`{
			"item_name": "%s"
		}`, itemName))

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		req, _ := http.NewRequest("", "", bytes.NewBuffer(jsonData))
		ctx.Params = gin.Params{
			{Key: "channelName", Value: channelName},
			{Key: "userID", Value: string(userID)},
		}
		ctx.Request = req

		return ctx
	}

	userID := models.TwitchID("user id")
	channelID := models.TwitchID("channel id")
	channelName := "channel name"
	itemName := "item name"

	itemID := uuid.New()
	image := "image"

	item := models.Item{ItemID: itemID, Image: image}

	announcerMock := mock.Mock[Announcer]()
	itemsMock := mock.Mock[ItemGetSetter]()
	viewersMock := mock.Mock[ViewerGetter]()
	usersMock := mock.Mock[UserIDGetter]()
	cacheMock := mock.Mock[ViewerCache]()

	mock.When(usersMock.GetUserID(channelName)).ThenReturn(channelID, nil)
	mock.When(itemsMock.GetItemByName(channelID, itemName)).ThenReturn(item, nil)

	controller := NewTwitchBotController(
		announcerMock,
		itemsMock,
		viewersMock,
		usersMock,
		cacheMock,
	)

	controller.UpdateViewer(setUpContext(userID, channelName, itemName))

	mock.Verify(itemsMock, mock.Once()).SetSelectedItem(userID, channelID, itemID)
	mock.Verify(cacheMock, mock.Once()).UpdateViewer(channelID, userID, image)
	mock.Verify(announcerMock, mock.Once()).AnnounceUpdate(channelName, image, userID)
}
