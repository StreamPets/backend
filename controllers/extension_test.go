package controllers

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
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/services"
)

func TestGetStoreData(t *testing.T) {
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

	channelId := models.TwitchId("channel id")
	userId := models.TwitchId("user id")
	tokenString := "token string"
	token := services.ExtToken{ChannelId: channelId, UserId: userId}

	storeItems := []models.Item{{}, {}}

	announcerMock := mock.Mock[UpdateAnnouncer]()
	verifierMock := mock.Mock[TokenVerifier]()
	storeMock := mock.Mock[StoreService]()
	usernameMock := mock.Mock[UsernameGetter]()

	mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(&token, nil)
	mock.When(storeMock.GetChannelsItems(channelId)).ThenReturn(storeItems, nil)

	controller := NewExtensionController(
		announcerMock,
		verifierMock,
		storeMock,
		usernameMock,
	)

	ctx, recorder := setUpContext(tokenString)
	controller.GetStoreData(ctx)

	mock.Verify(verifierMock, mock.Once()).VerifyExtToken(tokenString)
	mock.Verify(storeMock, mock.Once()).GetChannelsItems(channelId)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected %d got %d", http.StatusOK, recorder.Code)
	}

	var response []models.Item
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("could not parse json response")
	}

	if !slices.Equal(response, storeItems) {
		t.Errorf("expected %s got %s", storeItems, response)
	}
}

func TestGetUserData(t *testing.T) {
	mock.SetUp(t)

	type Response struct {
		OwnedItems   []models.Item `json:"owned"`
		SelectedItem models.Item   `json:"selected"`
	}

	setUpContext := func(tokenString string) (*gin.Context, *httptest.ResponseRecorder) {
		gin.SetMode(gin.TestMode)

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest("GET", "/items", nil)

		req.Header.Add("x-extension-jwt", tokenString)

		ctx.Request = req
		return ctx, recorder
	}

	channelId := models.TwitchId("channel id")
	userId := models.TwitchId("user id")

	tokenString := "token string"
	token := &services.ExtToken{
		UserId:    userId,
		ChannelId: channelId,
	}

	selectedItem := models.Item{ItemId: uuid.New()}
	ownedItems := []models.Item{selectedItem}

	announcerMock := mock.Mock[UpdateAnnouncer]()
	verifierMock := mock.Mock[TokenVerifier]()
	storeMock := mock.Mock[StoreService]()
	usernameMock := mock.Mock[UsernameGetter]()

	mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(token, nil)
	mock.When(storeMock.GetOwnedItems(channelId, userId)).ThenReturn(ownedItems, nil)
	mock.When(storeMock.GetSelectedItem(userId, channelId)).ThenReturn(selectedItem, nil)

	extController := NewExtensionController(
		announcerMock,
		verifierMock,
		storeMock,
		usernameMock,
	)

	ctx, recorder := setUpContext(tokenString)
	extController.GetUserData(ctx)

	var response Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("could not parse json response")
	}

	if !slices.Equal(response.OwnedItems, ownedItems) {
		t.Errorf("got %s want %s", response.OwnedItems, ownedItems)
	}
	if response.SelectedItem != selectedItem {
		t.Errorf("got %s want %s", response.SelectedItem, selectedItem)
	}
}

func TestBuyStoreItem(t *testing.T) {
	mock.SetUp(t)

	setUpContext := func(token, receipt, itemId string) *gin.Context {
		gin.SetMode(gin.TestMode)

		jsonData := []byte(fmt.Sprintf(`{
			"receipt": "%s",
			"item_id": "%s"
		}`, receipt, itemId))

		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		req, _ := http.NewRequest("POST", "/items", bytes.NewBuffer(jsonData))

		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
		req.Header.Add("x-extension-jwt", token)

		ctx.Request = req
		return ctx
	}

	channelId := models.TwitchId("channel id")
	userId := models.TwitchId("user id")
	itemId := uuid.New()
	transactionId := uuid.New()
	rarity := models.Rarity("common")

	tokenString := "token string"
	receiptString := "receipt string"

	token := services.ExtToken{ChannelId: channelId, UserId: userId}
	receipt := services.Receipt{
		Data: services.Data{
			TransactionId: transactionId,
			Product: services.Product{
				Rarity: rarity,
			},
		},
	}

	item := models.Item{ItemId: itemId, Rarity: rarity}

	announcerMock := mock.Mock[UpdateAnnouncer]()
	verifierMock := mock.Mock[TokenVerifier]()
	storeMock := mock.Mock[StoreService]()
	usernameMock := mock.Mock[UsernameGetter]()

	mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(&token, nil)
	mock.When(verifierMock.VerifyReceipt(receiptString)).ThenReturn(&receipt, nil)
	mock.When(storeMock.GetItemById(itemId)).ThenReturn(item, nil)
	mock.When(storeMock.AddOwnedItem(userId, itemId, transactionId)).ThenReturn(nil)

	controller := NewExtensionController(
		announcerMock,
		verifierMock,
		storeMock,
		usernameMock,
	)

	controller.BuyStoreItem(setUpContext(tokenString, receiptString, itemId.String()))

	mock.Verify(verifierMock, mock.Once()).VerifyExtToken(tokenString)
	mock.Verify(verifierMock, mock.Once()).VerifyReceipt(receiptString)
	mock.Verify(storeMock, mock.Once()).AddOwnedItem(userId, itemId, transactionId)
}

func TestSetSelectedItem(t *testing.T) {
	mock.SetUp(t)

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

	channelName := "channel name"
	tokenString := "token string"
	image := "image"

	channelId := models.TwitchId("channel id")
	userId := models.TwitchId("user id")
	itemId := uuid.New()

	item := models.Item{ItemId: itemId, Image: image}

	token := services.ExtToken{
		ChannelId: channelId,
		UserId:    userId,
	}

	announcerMock := mock.Mock[UpdateAnnouncer]()
	verifierMock := mock.Mock[TokenVerifier]()
	storeMock := mock.Mock[StoreService]()
	usernameMock := mock.Mock[UsernameGetter]()

	mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(&token, nil)
	mock.When(usernameMock.GetUsername(channelId)).ThenReturn(channelName, nil)
	mock.When(storeMock.SetSelectedItem(userId, channelId, itemId)).ThenReturn(nil)
	mock.When(storeMock.GetItemById(itemId)).ThenReturn(item, nil)

	controller := NewExtensionController(
		announcerMock,
		verifierMock,
		storeMock,
		usernameMock,
	)

	ctx := setUpContext(tokenString, itemId.String())

	controller.SetSelectedItem(ctx)

	mock.Verify(verifierMock, mock.Once()).VerifyExtToken(tokenString)
	mock.Verify(storeMock, mock.Once()).SetSelectedItem(userId, channelId, itemId)
	mock.Verify(announcerMock, mock.Once()).AnnounceUpdate(channelName, image, userId)
}
