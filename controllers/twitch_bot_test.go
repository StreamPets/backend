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
	databaseMock := mock.Mock[ViewerItemSetterGetter]()
	usersMock := mock.Mock[UserIDGetter]()

	mock.When(usersMock.GetUserID(channelName)).ThenReturn(channelID, nil)
	mock.When(databaseMock.GetViewer(userID, channelID, username)).ThenReturn(viewer, nil)

	controller := NewTwitchBotController(
		announcerMock,
		databaseMock,
		usersMock,
	)

	controller.AddViewerToChannel(setUpContext(userID, channelName, username))

	mock.Verify(announcerMock, mock.Once()).AnnounceJoin(channelName, viewer)
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
	channelName := "channel name"

	announcerMock := mock.Mock[Announcer]()
	databaseMock := mock.Mock[ViewerItemSetterGetter]()
	usersMock := mock.Mock[UserIDGetter]()

	controller := NewTwitchBotController(
		announcerMock,
		databaseMock,
		usersMock,
	)

	controller.RemoveViewerFromChannel(setUpContext(userID, channelName))

	mock.Verify(announcerMock, mock.Once()).AnnouncePart(channelName, userID)
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
	databaseMock := mock.Mock[ViewerItemSetterGetter]()
	usersMock := mock.Mock[UserIDGetter]()

	controller := NewTwitchBotController(
		announcerMock,
		databaseMock,
		usersMock,
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
	databaseMock := mock.Mock[ViewerItemSetterGetter]()
	usersMock := mock.Mock[UserIDGetter]()

	mock.When(usersMock.GetUserID(channelName)).ThenReturn(channelID, nil)
	mock.When(databaseMock.GetItemByName(channelID, itemName)).ThenReturn(item, nil)

	controller := NewTwitchBotController(
		announcerMock,
		databaseMock,
		usersMock,
	)

	controller.UpdateViewer(setUpContext(userID, channelName, itemName))

	mock.Verify(databaseMock, mock.Once()).SetSelectedItem(userID, channelID, itemID)
	mock.Verify(announcerMock, mock.Once()).AnnounceUpdate(channelName, image, userID)
}
