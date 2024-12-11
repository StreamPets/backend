package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/controllers"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/services"
)

func TestGetStoreData(t *testing.T) {
	type Response struct {
		Owned []models.Item `json:"owned"`
		Store []models.Item `json:"store"`
	}

	mock.SetUp(t)

	setUpContext := func(tokenString string) (*gin.Context, *httptest.ResponseRecorder) {
		gin.SetMode(gin.TestMode)

		recorder := httptest.NewRecorder()
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

	storeItems := []models.Item{{}, {}}
	ownedItems := []models.Item{{}}

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

	if !slices.Equal(response.Owned, ownedItems) {
		t.Errorf("expected %s got %s", ownedItems, response.Owned)
	}
	if !slices.Equal(response.Store, storeItems) {
		t.Errorf("expected %s got %s", storeItems, response.Store)
	}
}

func TestBuyStoreItem(t *testing.T) {
	mock.SetUp(t)

	setUpContext := func(token, receipt, itemID string) *gin.Context {
		gin.SetMode(gin.TestMode)

		jsonData := []byte(fmt.Sprintf(`{
			"receipt": "%s",
			"item_id": "%s"
		}`, receipt, itemID))

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		req, _ := http.NewRequest("POST", "/items", bytes.NewBuffer(jsonData))

		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
		req.Header.Add("x-extension-jwt", token)

		ctx.Request = req
		return ctx
	}

	channelID := models.TwitchID("channel id")
	userID := models.TwitchID("user id")
	itemID := uuid.New()
	transactionID := uuid.New()

	tokenString := "token string"
	receiptString := "receipt string"

	token := services.ExtToken{ChannelID: channelID, UserID: userID}
	receipt := services.Receipt{TransactionID: transactionID}

	announcerMock := mock.Mock[services.Announcer]()
	authMock := mock.Mock[services.AuthService]()
	twitchMock := mock.Mock[repositories.TwitchRepository]()
	databaseMock := mock.Mock[services.DatabaseService]()

	mock.When(authMock.VerifyExtToken(tokenString)).ThenReturn(&token, nil)
	mock.When(authMock.VerifyReceipt(receiptString)).ThenReturn(&receipt, nil)
	mock.When(databaseMock.AddOwnedItem(userID, itemID, transactionID)).ThenReturn(nil)

	controller := controllers.NewController(
		announcerMock,
		authMock,
		twitchMock,
		databaseMock,
	)

	controller.BuyStoreItem(setUpContext(tokenString, receiptString, itemID.String()))

	mock.Verify(authMock, mock.Once()).VerifyExtToken(tokenString)
	mock.Verify(authMock, mock.Once()).VerifyReceipt(receiptString)
	mock.Verify(databaseMock, mock.Once()).AddOwnedItem(userID, itemID, transactionID)
}

func TestSetSelectedItem(t *testing.T) {
	mock.SetUp(t)

	setUpContext := func(token, itemID string) *gin.Context {
		gin.SetMode(gin.TestMode)

		jsonData := []byte(fmt.Sprintf(`{
			"item_id": "%s"
		}`, itemID))

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		req, _ := http.NewRequest("POST", "/items", bytes.NewBuffer(jsonData))

		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
		req.Header.Add("x-extension-jwt", token)

		ctx.Request = req
		return ctx
	}

	channelName := "channel name"
	tokenString := "token string"
	image := "image"

	channelID := models.TwitchID("channel id")
	userID := models.TwitchID("user id")
	itemID := uuid.New()

	item := models.Item{ItemID: itemID, Image: image}

	token := services.ExtToken{
		ChannelID: channelID,
		UserID:    userID,
	}

	announcerMock := mock.Mock[services.Announcer]()
	authMock := mock.Mock[services.AuthService]()
	twitchMock := mock.Mock[repositories.TwitchRepository]()
	databaseMock := mock.Mock[services.DatabaseService]()

	mock.When(authMock.VerifyExtToken(tokenString)).ThenReturn(&token, nil)
	mock.When(twitchMock.GetUsername(channelID)).ThenReturn(channelName, nil)
	mock.When(databaseMock.SetSelectedItem(userID, channelID, itemID)).ThenReturn(nil)
	mock.When(databaseMock.GetItemByID(itemID)).ThenReturn(item, nil)

	controller := controllers.NewController(
		announcerMock,
		authMock,
		twitchMock,
		databaseMock,
	)

	ctx := setUpContext(tokenString, itemID.String())

	controller.SetSelectedItem(ctx)

	mock.Verify(authMock, mock.Once()).VerifyExtToken(tokenString)
	mock.Verify(databaseMock, mock.Once()).SetSelectedItem(userID, channelID, itemID)
	mock.Verify(announcerMock, mock.Once()).AnnounceUpdate(channelName, image, userID)
}
