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

	announcerMock := mock.Mock[UpdateAnnouncer]()
	verifierMock := mock.Mock[TokenVerifier]()
	storeMock := mock.Mock[StoreService]()
	usersMock := mock.Mock[UserGetter]()

	mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(&token, nil)
	mock.When(storeMock.GetTodaysItems(channelID)).ThenReturn(storeItems, nil)
	mock.When(storeMock.GetOwnedItems(channelID, userID)).ThenReturn(ownedItems, nil)

	controller := NewExtensionController(
		announcerMock,
		verifierMock,
		storeMock,
		usersMock,
	)

	ctx, recorder := setUpContext(tokenString)
	controller.GetStoreData(ctx)

	mock.Verify(verifierMock, mock.Once()).VerifyExtToken(tokenString)
	mock.Verify(storeMock, mock.Once()).GetTodaysItems(channelID)
	mock.Verify(storeMock, mock.Once()).GetOwnedItems(channelID, userID)

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

	announcerMock := mock.Mock[UpdateAnnouncer]()
	verifierMock := mock.Mock[TokenVerifier]()
	storeMock := mock.Mock[StoreService]()
	usersMock := mock.Mock[UserGetter]()

	mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(&token, nil)
	mock.When(verifierMock.VerifyReceipt(receiptString)).ThenReturn(&receipt, nil)
	mock.When(storeMock.AddOwnedItem(userID, itemID, transactionID)).ThenReturn(nil)

	controller := NewExtensionController(
		announcerMock,
		verifierMock,
		storeMock,
		usersMock,
	)

	controller.BuyStoreItem(setUpContext(tokenString, receiptString, itemID.String()))

	mock.Verify(verifierMock, mock.Once()).VerifyExtToken(tokenString)
	mock.Verify(verifierMock, mock.Once()).VerifyReceipt(receiptString)
	mock.Verify(storeMock, mock.Once()).AddOwnedItem(userID, itemID, transactionID)
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

	announcerMock := mock.Mock[UpdateAnnouncer]()
	verifierMock := mock.Mock[TokenVerifier]()
	storeMock := mock.Mock[StoreService]()
	usersMock := mock.Mock[UserGetter]()

	mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(&token, nil)
	mock.When(usersMock.GetUsername(channelID)).ThenReturn(channelName, nil)
	mock.When(storeMock.SetSelectedItem(userID, channelID, itemID)).ThenReturn(nil)
	mock.When(storeMock.GetItemByID(itemID)).ThenReturn(item, nil)

	controller := NewExtensionController(
		announcerMock,
		verifierMock,
		storeMock,
		usersMock,
	)

	ctx := setUpContext(tokenString, itemID.String())

	controller.SetSelectedItem(ctx)

	mock.Verify(verifierMock, mock.Once()).VerifyExtToken(tokenString)
	mock.Verify(storeMock, mock.Once()).SetSelectedItem(userID, channelID, itemID)
	mock.Verify(announcerMock, mock.Once()).AnnounceUpdate(channelName, image, userID)
}
