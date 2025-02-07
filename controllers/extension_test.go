package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
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

var ErrTestError error = errors.New("this is a test error")

func TestGetStoreData(t *testing.T) {
	setUpContext := func(tokenString string) (*gin.Context, *httptest.ResponseRecorder) {
		gin.SetMode(gin.TestMode)

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest("GET", "/items", nil)

		req.Header.Add("x-extension-jwt", tokenString)

		ctx.Request = req
		return ctx, recorder
	}

	t.Run("access forbidden with invalid extension token", func(t *testing.T) {
		mock.SetUp(t)

		tokenString := "invalid token"

		announcerMock := mock.Mock[UpdateAnnouncer]()
		verifierMock := mock.Mock[TokenVerifier]()
		storeMock := mock.Mock[StoreService]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(nil, services.ErrInvalidToken)

		controller := NewExtensionController(
			announcerMock,
			verifierMock,
			storeMock,
		)

		ctx, recorder := setUpContext(tokenString)
		controller.GetStoreData(ctx)

		assert.NotEqual(t, recorder.Code, http.StatusOK)
	})

	t.Run("not found with invalid channel id", func(t *testing.T) {
		mock.SetUp(t)

		tokenString := "extension token"
		channelId := twitch.Id("channel id")

		token := &services.ExtToken{
			ChannelId: channelId,
		}

		announcerMock := mock.Mock[UpdateAnnouncer]()
		verifierMock := mock.Mock[TokenVerifier]()
		storeMock := mock.Mock[StoreService]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(token, nil)
		mock.When(storeMock.GetChannelsItems(channelId)).ThenReturn(nil, ErrTestError)

		controller := NewExtensionController(
			announcerMock,
			verifierMock,
			storeMock,
		)

		ctx, recorder := setUpContext(tokenString)
		controller.GetStoreData(ctx)

		assert.NotEqual(t, recorder.Code, http.StatusOK)
	})

	t.Run("items returned when extension token and channel id are valid", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.Id("channel id")
		userId := twitch.Id("user id")
		tokenString := "token string"
		token := services.ExtToken{ChannelId: channelId, UserId: userId}

		storeItems := []models.Item{{}, {}}

		announcerMock := mock.Mock[UpdateAnnouncer]()
		verifierMock := mock.Mock[TokenVerifier]()
		storeMock := mock.Mock[StoreService]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(&token, nil)
		mock.When(storeMock.GetChannelsItems(channelId)).ThenReturn(storeItems, nil)

		controller := NewExtensionController(
			announcerMock,
			verifierMock,
			storeMock,
		)

		ctx, recorder := setUpContext(tokenString)
		controller.GetStoreData(ctx)

		mock.Verify(verifierMock, mock.Once()).VerifyExtToken(tokenString)
		mock.Verify(storeMock, mock.Once()).GetChannelsItems(channelId)

		assert.Equal(t, recorder.Code, http.StatusOK)

		var response []models.Item
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Errorf("could not parse json response")
		}

		assert.Equal(t, storeItems, response)
	})
}

func TestGetUserData(t *testing.T) {
	setUpContext := func(tokenString string) (*gin.Context, *httptest.ResponseRecorder) {
		gin.SetMode(gin.TestMode)

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest("GET", "/items", nil)

		req.Header.Add("x-extension-jwt", tokenString)

		ctx.Request = req
		return ctx, recorder
	}

	t.Run("error when extension token is invalid", func(t *testing.T) {
		mock.SetUp(t)

		tokenString := "token string"

		announcerMock := mock.Mock[UpdateAnnouncer]()
		verifierMock := mock.Mock[TokenVerifier]()
		storeMock := mock.Mock[StoreService]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(nil, services.ErrInvalidToken)

		extController := NewExtensionController(
			announcerMock,
			verifierMock,
			storeMock,
		)

		ctx, recorder := setUpContext(tokenString)
		extController.GetUserData(ctx)

		assert.NotEqual(t, recorder.Code, http.StatusOK)
	})

	t.Run("fail when get owned items fails", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.Id("channel id")
		userId := twitch.Id("user id")

		tokenString := "token string"
		token := &services.ExtToken{
			ChannelId: channelId,
			UserId:    userId,
		}

		announcerMock := mock.Mock[UpdateAnnouncer]()
		verifierMock := mock.Mock[TokenVerifier]()
		storeMock := mock.Mock[StoreService]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(token, nil)
		mock.When(storeMock.GetOwnedItems(channelId, userId)).ThenReturn(nil, ErrTestError)

		extController := NewExtensionController(
			announcerMock,
			verifierMock,
			storeMock,
		)

		ctx, recorder := setUpContext(tokenString)
		extController.GetUserData(ctx)

		assert.NotEqual(t, recorder.Code, http.StatusOK)
	})

	t.Run("fail when get selected items fails", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.Id("channel id")
		userId := twitch.Id("user id")

		tokenString := "token string"
		token := &services.ExtToken{
			ChannelId: channelId,
			UserId:    userId,
		}

		announcerMock := mock.Mock[UpdateAnnouncer]()
		verifierMock := mock.Mock[TokenVerifier]()
		storeMock := mock.Mock[StoreService]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(token, nil)
		mock.When(storeMock.GetSelectedItem(userId, channelId)).ThenReturn(nil, ErrTestError)

		extController := NewExtensionController(
			announcerMock,
			verifierMock,
			storeMock,
		)

		ctx, recorder := setUpContext(tokenString)
		extController.GetUserData(ctx)

		assert.NotEqual(t, recorder.Code, http.StatusOK)
	})

	t.Run("status ok when all pre-requisites are met", func(t *testing.T) {
		mock.SetUp(t)

		type Response struct {
			OwnedItems   []models.Item `json:"owned"`
			SelectedItem models.Item   `json:"selected"`
		}

		channelId := twitch.Id("channel id")
		userId := twitch.Id("user id")

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

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(token, nil)
		mock.When(storeMock.GetOwnedItems(channelId, userId)).ThenReturn(ownedItems, nil)
		mock.When(storeMock.GetSelectedItem(userId, channelId)).ThenReturn(selectedItem, nil)

		extController := NewExtensionController(
			announcerMock,
			verifierMock,
			storeMock,
		)

		ctx, recorder := setUpContext(tokenString)
		extController.GetUserData(ctx)

		var response Response
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Errorf("could not parse json response")
		}

		assert.Equal(t, response.OwnedItems, ownedItems)
		assert.Equal(t, response.SelectedItem, selectedItem)
	})
}

func TestBuyStoreItem(t *testing.T) {
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

	t.Run("item not added when extension token is invalid", func(t *testing.T) {
		mock.SetUp(t)

		userId := twitch.Id("user id")

		itemId := uuid.New()
		transactionId := uuid.New()

		tokenString := "token string"
		receiptString := "receipt string"

		announcerMock := mock.Mock[UpdateAnnouncer]()
		verifierMock := mock.Mock[TokenVerifier]()
		storeMock := mock.Mock[StoreService]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(nil, services.ErrInvalidToken)

		extController := NewExtensionController(
			announcerMock,
			verifierMock,
			storeMock,
		)

		extController.BuyStoreItem(setUpContext(tokenString, receiptString, itemId.String()))

		mock.Verify(storeMock, mock.Never()).AddOwnedItem(userId, itemId, transactionId)
	})

	t.Run("item not added when item id is not a valid uuid", func(t *testing.T) {
		mock.SetUp(t)

		itemId := "invalid id"

		tokenString := "token string"
		receiptString := "receipt string"

		announcerMock := mock.Mock[UpdateAnnouncer]()
		verifierMock := mock.Mock[TokenVerifier]()
		storeMock := mock.Mock[StoreService]()

		extController := NewExtensionController(
			announcerMock,
			verifierMock,
			storeMock,
		)

		extController.BuyStoreItem(setUpContext(tokenString, receiptString, itemId))
	})

	t.Run("item not added when item id does not exist", func(t *testing.T) {
		mock.SetUp(t)

		userId := twitch.Id("user id")

		itemId := uuid.New()
		transactionId := uuid.New()

		tokenString := "token string"
		receiptString := "receipt string"

		announcerMock := mock.Mock[UpdateAnnouncer]()
		verifierMock := mock.Mock[TokenVerifier]()
		storeMock := mock.Mock[StoreService]()

		mock.When(storeMock.GetItemById(itemId)).ThenReturn(nil, ErrTestError)

		extController := NewExtensionController(
			announcerMock,
			verifierMock,
			storeMock,
		)

		extController.BuyStoreItem(setUpContext(tokenString, receiptString, itemId.String()))

		mock.Verify(storeMock, mock.Never()).AddOwnedItem(userId, itemId, transactionId)
	})

	t.Run("item not added when receipt is invalid", func(t *testing.T) {
		mock.SetUp(t)

		userId := twitch.Id("user id")

		itemId := uuid.New()
		transactionId := uuid.New()

		tokenString := "token string"
		receiptString := "receipt string"

		announcerMock := mock.Mock[UpdateAnnouncer]()
		verifierMock := mock.Mock[TokenVerifier]()
		storeMock := mock.Mock[StoreService]()

		mock.When(verifierMock.VerifyReceipt(receiptString)).ThenReturn(nil, services.ErrInvalidToken)

		extController := NewExtensionController(
			announcerMock,
			verifierMock,
			storeMock,
		)

		extController.BuyStoreItem(setUpContext(tokenString, receiptString, itemId.String()))

		mock.Verify(storeMock, mock.Never()).AddOwnedItem(userId, itemId, transactionId)
	})

	t.Run("item not added when receipt and item rarity do not match", func(t *testing.T) {
		mock.SetUp(t)

		tokenString := "token string"
		receiptString := "receipt string"

		userId := twitch.Id("user id")

		itemId := uuid.New()
		transactionId := uuid.New()

		item := models.Item{
			ItemId: itemId,
			Rarity: models.Common,
		}

		receipt := &services.Receipt{
			Data: services.Data{
				TransactionId: transactionId,
				Product: services.Product{
					Rarity: models.Uncommon,
				},
			},
		}

		announcerMock := mock.Mock[UpdateAnnouncer]()
		verifierMock := mock.Mock[TokenVerifier]()
		storeMock := mock.Mock[StoreService]()

		mock.When(storeMock.GetItemById(itemId)).ThenReturn(item, nil)
		mock.When(verifierMock.VerifyReceipt(receiptString)).ThenReturn(receipt, nil)

		extController := NewExtensionController(
			announcerMock,
			verifierMock,
			storeMock,
		)

		extController.BuyStoreItem(setUpContext(tokenString, receiptString, itemId.String()))

		mock.Verify(storeMock, mock.Never()).AddOwnedItem(userId, itemId, transactionId)
	})

	t.Run("item added when all pre-requisites are met", func(t *testing.T) {
		mock.SetUp(t)

		tokenString := "token string"
		receiptString := "receipt string"

		userId := twitch.Id("user id")

		itemId := uuid.New()
		transactionId := uuid.New()

		token := &services.ExtToken{
			UserId: userId,
		}

		receipt := &services.Receipt{
			Data: services.Data{
				TransactionId: transactionId,
				Product: services.Product{
					Rarity: models.Common,
				},
			},
		}

		item := models.Item{
			ItemId: itemId,
			Rarity: models.Common,
		}

		announcerMock := mock.Mock[UpdateAnnouncer]()
		verifierMock := mock.Mock[TokenVerifier]()
		storeMock := mock.Mock[StoreService]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(token, nil)
		mock.When(verifierMock.VerifyReceipt(receiptString)).ThenReturn(receipt, nil)
		mock.When(storeMock.GetItemById(itemId)).ThenReturn(item, nil)

		extController := NewExtensionController(
			announcerMock,
			verifierMock,
			storeMock,
		)

		extController.BuyStoreItem(setUpContext(tokenString, receiptString, itemId.String()))

		mock.Verify(storeMock, mock.Once()).AddOwnedItem(userId, itemId, transactionId)
	})
}

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
		verifierMock := mock.Mock[TokenVerifier]()
		storeMock := mock.Mock[StoreService]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(nil, ErrTestError)

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
		verifierMock := mock.Mock[TokenVerifier]()
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
		verifierMock := mock.Mock[TokenVerifier]()
		storeMock := mock.Mock[StoreService]()

		mock.When(storeMock.GetItemById(itemId)).ThenReturn(nil, ErrTestError)

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
		verifierMock := mock.Mock[TokenVerifier]()
		storeMock := mock.Mock[StoreService]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(token, nil)
		mock.When(storeMock.SetSelectedItem(userId, channelId, itemId)).ThenReturn(ErrTestError)

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
		verifierMock := mock.Mock[TokenVerifier]()
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
