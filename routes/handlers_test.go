package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ovechkin-dm/mockio/mock"
	"github.com/streampets/backend/announcers"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/services"
	"github.com/streampets/backend/test"
	"github.com/streampets/backend/twitch"
	"github.com/stretchr/testify/assert"
)

func TestHandleLogin(t *testing.T) {
	setUpContext := func(cookie string) (*gin.Context, *httptest.ResponseRecorder) {
		gin.SetMode(gin.TestMode)

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest("GET", "/items", nil)

		if cookie != "" {
			req.AddCookie(&http.Cookie{
				Name:  "Authorization",
				Value: cookie,
			})
		}

		ctx.Request = req
		return ctx, recorder
	}

	t.Run("unauthorized status when no 'Authorization' cookie present", func(t *testing.T) {
		mock.SetUp(t)

		overlays := mock.Mock[overlayIdGetter]()
		validator := mock.Mock[tokenValidator]()

		ctx, recorder := setUpContext("")
		handleLogin(validator, overlays)(ctx)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run("unauthorized status when access token invalid", func(t *testing.T) {
		mock.SetUp(t)

		invalidToken := "inavlid token"
		ctx, recorder := setUpContext(invalidToken)

		overlays := mock.Mock[overlayIdGetter]()
		validator := mock.Mock[tokenValidator]()

		mock.When(validator.ValidateToken(ctx, invalidToken)).ThenReturn(nil, twitch.ErrInvalidUserToken)

		handleLogin(validator, overlays)(ctx)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run("internal server error when validate token fails", func(t *testing.T) {
		mock.SetUp(t)

		invalidToken := "inavlid token"
		ctx, recorder := setUpContext(invalidToken)

		overlays := mock.Mock[overlayIdGetter]()
		validator := mock.Mock[tokenValidator]()

		mock.When(validator.ValidateToken(ctx, invalidToken)).ThenReturn(nil, assert.AnError)

		handleLogin(validator, overlays)(ctx)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("status bad request when channel id has no overlay id", func(t *testing.T) {
		mock.SetUp(t)

		token := "token"
		channelId := twitch.Id("channel id")
		ctx, recorder := setUpContext(token)

		overlays := mock.Mock[overlayIdGetter]()
		validator := mock.Mock[tokenValidator]()

		mock.When(validator.ValidateToken(ctx, token)).ThenReturn(channelId, nil)
		mock.When(overlays.GetOverlayId(channelId)).ThenReturn(nil, repositories.NewErrNoOverlayId(channelId))

		handleLogin(validator, overlays)(ctx)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("status bad request when channel id has no overlay id", func(t *testing.T) {
		mock.SetUp(t)

		token := "token"
		channelId := twitch.Id("channel id")
		ctx, recorder := setUpContext(token)

		overlays := mock.Mock[overlayIdGetter]()
		validator := mock.Mock[tokenValidator]()

		mock.When(validator.ValidateToken(ctx, token)).ThenReturn(channelId, nil)
		mock.When(overlays.GetOverlayId(channelId)).ThenReturn(nil, assert.AnError)

		handleLogin(validator, overlays)(ctx)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("channel id and overlay id returned in normal case", func(t *testing.T) {
		mock.SetUp(t)

		type userData struct {
			OverlayId uuid.UUID `json:"overlay_id"`
			ChannelId twitch.Id `json:"channel_id"`
		}

		token := "token"
		channelId := twitch.Id("channel id")
		overlayId := uuid.New()
		ctx, recorder := setUpContext(token)

		overlays := mock.Mock[overlayIdGetter]()
		validator := mock.Mock[tokenValidator]()

		mock.When(validator.ValidateToken(ctx, token)).ThenReturn(channelId, nil)
		mock.When(overlays.GetOverlayId(channelId)).ThenReturn(overlayId, nil)

		handleLogin(validator, overlays)(ctx)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var actual userData
		if err := json.Unmarshal(recorder.Body.Bytes(), &actual); err != nil {
			t.Errorf("could not parse json response")
		}

		expected := userData{
			OverlayId: overlayId,
			ChannelId: channelId,
		}

		assert.Equal(t, expected, actual)
	})
}

func TestHandleListen(t *testing.T) {
	setUpContext := func(channelId twitch.Id, overlayId uuid.UUID) (*gin.Context, *test.CloseNotifierResponseWriter) {
		gin.SetMode(gin.TestMode)

		recorder := &test.CloseNotifierResponseWriter{ResponseRecorder: httptest.NewRecorder()}
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest("GET", "/listen", nil)

		values := req.URL.Query()
		values.Add("channelId", string(channelId))
		values.Add("overlayId", overlayId.String())
		req.URL.RawQuery = values.Encode()

		ctx.Request = req
		return ctx, recorder
	}

	channelId := twitch.Id("channel id")
	overlayId := uuid.New()

	t.Run("receive and send events from stream", func(t *testing.T) {
		mock.SetUp(t)

		ctx, recorder := setUpContext(channelId, overlayId)

		stream := make(chan announcers.Announcement)
		client := announcers.Client{Stream: stream}

		announcerMock := mock.Mock[clientAddRemover]()
		authMock := mock.Mock[overlayIdValidator]()

		mock.When(announcerMock.AddClient(channelId)).ThenReturn(client)

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			handleListen(announcerMock, authMock)(ctx)
		}()

		stream <- announcers.Announcement{
			Event:   "event",
			Message: "message",
		}

		close(stream)
		wg.Wait()

		mock.Verify(authMock, mock.Once()).ValidateOverlayId(channelId, overlayId)
		mock.Verify(announcerMock, mock.Once()).AddClient(channelId)
		mock.Verify(announcerMock, mock.Once()).RemoveClient(client)

		assert.Contains(t, recorder.Body.String(), "event:event")
		assert.Contains(t, recorder.Body.String(), "data:message")
	})

	t.Run("client not added when overlay id and channel id do not match", func(t *testing.T) {
		mock.SetUp(t)

		ctx, recorder := setUpContext(channelId, overlayId)

		announcerMock := mock.Mock[clientAddRemover]()
		authMock := mock.Mock[overlayIdValidator]()

		mock.When(authMock.ValidateOverlayId(channelId, overlayId)).ThenReturn(services.ErrIdMismatch)

		handleListen(announcerMock, authMock)(ctx)

		mock.Verify(authMock, mock.Once()).ValidateOverlayId(channelId, overlayId)
		mock.Verify(announcerMock, mock.Never()).AddClient(channelId)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})
}

func TestGetStoreData(t *testing.T) {
	setUpContext := func(tokenString string) (*gin.Context, *httptest.ResponseRecorder) {
		gin.SetMode(gin.TestMode)

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest("GET", "/items", nil)

		req.Header.Add(XExtensionJwt, tokenString)

		ctx.Request = req
		return ctx, recorder
	}

	t.Run("access forbidden with invalid extension token", func(t *testing.T) {
		mock.SetUp(t)

		tokenString := "invalid token"

		verifierMock := mock.Mock[extTokenVerifier]()
		storeMock := mock.Mock[channelItemGetter]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(nil, services.NewErrInvalidToken(tokenString))

		ctx, recorder := setUpContext(tokenString)
		handleGetStoreData(verifierMock, storeMock)(ctx)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run("internal server error when error received from verify token", func(t *testing.T) {
		mock.SetUp(t)

		tokenString := "invalid token"

		verifierMock := mock.Mock[extTokenVerifier]()
		storeMock := mock.Mock[channelItemGetter]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(nil, assert.AnError)

		ctx, recorder := setUpContext(tokenString)
		handleGetStoreData(verifierMock, storeMock)(ctx)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("internal server error when error received from get channels items", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.Id("channel id")
		userId := twitch.Id("user id")
		tokenString := "token string"
		token := services.ExtToken{ChannelId: channelId, UserId: userId}

		storeItems := []models.Item{{}, {}}

		verifierMock := mock.Mock[extTokenVerifier]()
		storeMock := mock.Mock[channelItemGetter]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(&token, nil)
		mock.When(storeMock.GetChannelsItems(channelId)).ThenReturn(storeItems, assert.AnError)

		ctx, recorder := setUpContext(tokenString)
		handleGetStoreData(verifierMock, storeMock)(ctx)

		mock.Verify(verifierMock, mock.Once()).VerifyExtToken(tokenString)
		mock.Verify(storeMock, mock.Once()).GetChannelsItems(channelId)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("items returned when extension token and channel id are valid", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.Id("channel id")
		userId := twitch.Id("user id")
		tokenString := "token string"
		token := services.ExtToken{ChannelId: channelId, UserId: userId}

		storeItems := []models.Item{{}, {}}

		verifierMock := mock.Mock[extTokenVerifier]()
		storeMock := mock.Mock[channelItemGetter]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(&token, nil)
		mock.When(storeMock.GetChannelsItems(channelId)).ThenReturn(storeItems, nil)

		ctx, recorder := setUpContext(tokenString)
		handleGetStoreData(verifierMock, storeMock)(ctx)

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

		req.Header.Add(XExtensionJwt, tokenString)

		ctx.Request = req
		return ctx, recorder
	}

	t.Run("unauthorized when token is invalid", func(t *testing.T) {
		mock.SetUp(t)

		tokenString := "token string"

		verifierMock := mock.Mock[extTokenVerifier]()
		storeMock := mock.Mock[userDataGetter]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(nil, services.NewErrInvalidToken(tokenString))

		ctx, recorder := setUpContext(tokenString)
		handleGetUserData(verifierMock, storeMock)(ctx)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run("internal server error when verify token fails", func(t *testing.T) {
		mock.SetUp(t)

		tokenString := "token string"

		verifierMock := mock.Mock[extTokenVerifier]()
		storeMock := mock.Mock[userDataGetter]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(nil, assert.AnError)

		ctx, recorder := setUpContext(tokenString)
		handleGetUserData(verifierMock, storeMock)(ctx)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("internal server error when get owned items fails", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.Id("channel id")
		userId := twitch.Id("user id")

		tokenString := "token string"
		token := &services.ExtToken{
			ChannelId: channelId,
			UserId:    userId,
		}

		verifierMock := mock.Mock[extTokenVerifier]()
		storeMock := mock.Mock[userDataGetter]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(token, nil)
		mock.When(storeMock.GetOwnedItems(channelId, userId)).ThenReturn(nil, assert.AnError)

		ctx, recorder := setUpContext(tokenString)
		handleGetUserData(verifierMock, storeMock)(ctx)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("internal server error when get selected item fails", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.Id("channel id")
		userId := twitch.Id("user id")

		tokenString := "token string"
		token := &services.ExtToken{
			ChannelId: channelId,
			UserId:    userId,
		}

		verifierMock := mock.Mock[extTokenVerifier]()
		storeMock := mock.Mock[userDataGetter]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(token, nil)
		mock.When(storeMock.GetSelectedItem(userId, channelId)).ThenReturn(nil, assert.AnError)

		ctx, recorder := setUpContext(tokenString)
		handleGetUserData(verifierMock, storeMock)(ctx)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
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

		verifierMock := mock.Mock[extTokenVerifier]()
		storeMock := mock.Mock[userDataGetter]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(token, nil)
		mock.When(storeMock.GetOwnedItems(channelId, userId)).ThenReturn(ownedItems, nil)
		mock.When(storeMock.GetSelectedItem(userId, channelId)).ThenReturn(selectedItem, nil)

		ctx, recorder := setUpContext(tokenString)
		handleGetUserData(verifierMock, storeMock)(ctx)

		assert.Equal(t, http.StatusOK, recorder.Code)

		var response Response
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Errorf("could not parse json response")
		}

		assert.Equal(t, response.OwnedItems, ownedItems)
		assert.Equal(t, response.SelectedItem, selectedItem)
	})
}

func TestBuyStoreItem(t *testing.T) {

	generateData := func(receipt, itemId string) []byte {
		return []byte(fmt.Sprintf(`{
			"receipt": "%s",
			"item_id": "%s"
		}`, receipt, itemId))
	}

	setUpContext := func(token string, jsonData []byte) (*gin.Context, *httptest.ResponseRecorder) {
		gin.SetMode(gin.TestMode)

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest("POST", "/items", bytes.NewBuffer(jsonData))

		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
		req.Header.Add(XExtensionJwt, token)

		ctx.Request = req
		return ctx, recorder
	}

	t.Run("item not added when extension token is invalid", func(t *testing.T) {
		mock.SetUp(t)

		itemId := uuid.New()

		tokenString := "token string"
		receiptString := "receipt string"

		verifierMock := mock.Mock[tokenVerifier]()
		storeMock := mock.Mock[foo]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(nil, services.NewErrInvalidToken(tokenString))

		jsonData := generateData(receiptString, itemId.String())
		ctx, recorder := setUpContext(tokenString, jsonData)
		handleBuyStoreItem(verifierMock, storeMock)(ctx)

		mock.VerifyNoMoreInteractions(storeMock)
		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run("item not added when extension token cannot be validated", func(t *testing.T) {
		mock.SetUp(t)

		itemId := uuid.New()

		tokenString := "token string"
		receiptString := "receipt string"

		verifierMock := mock.Mock[tokenVerifier]()
		storeMock := mock.Mock[foo]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(nil, assert.AnError)

		jsonData := generateData(receiptString, itemId.String())
		ctx, recorder := setUpContext(tokenString, jsonData)
		handleBuyStoreItem(verifierMock, storeMock)(ctx)

		mock.VerifyNoMoreInteractions(storeMock)
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("bad request when json is invalid format", func(t *testing.T) {
		mock.SetUp(t)

		tokenString := "token string"

		verifierMock := mock.Mock[tokenVerifier]()
		storeMock := mock.Mock[foo]()

		ctx, recorder := setUpContext(tokenString, []byte{})
		handleBuyStoreItem(verifierMock, storeMock)(ctx)

		mock.VerifyNoMoreInteractions(storeMock)
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("item not added when receipt is invalid", func(t *testing.T) {
		mock.SetUp(t)

		itemId := uuid.New()

		tokenString := "token string"
		receiptString := "receipt string"

		verifierMock := mock.Mock[tokenVerifier]()
		storeMock := mock.Mock[foo]()

		mock.When(verifierMock.VerifyReceipt(receiptString)).ThenReturn(nil, services.NewErrInvalidToken(receiptString))

		jsonData := generateData(receiptString, itemId.String())
		ctx, recorder := setUpContext(tokenString, jsonData)
		handleBuyStoreItem(verifierMock, storeMock)(ctx)

		mock.VerifyNoMoreInteractions(storeMock)
		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run("item not added when receipt cannot be validated", func(t *testing.T) {
		mock.SetUp(t)

		itemId := uuid.New()

		tokenString := "token string"
		receiptString := "receipt string"

		verifierMock := mock.Mock[tokenVerifier]()
		storeMock := mock.Mock[foo]()

		mock.When(verifierMock.VerifyReceipt(receiptString)).ThenReturn(nil, assert.AnError)

		jsonData := generateData(receiptString, itemId.String())
		ctx, recorder := setUpContext(tokenString, jsonData)
		handleBuyStoreItem(verifierMock, storeMock)(ctx)

		mock.VerifyNoMoreInteractions(storeMock)
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("item not added when item id is not a valid uuid", func(t *testing.T) {
		mock.SetUp(t)

		itemId := "invalid id"

		tokenString := "token string"
		receiptString := "receipt string"

		verifierMock := mock.Mock[tokenVerifier]()
		storeMock := mock.Mock[foo]()

		jsonData := generateData(receiptString, itemId)
		ctx, recorder := setUpContext(tokenString, jsonData)
		handleBuyStoreItem(verifierMock, storeMock)(ctx)

		mock.VerifyNoMoreInteractions(storeMock)
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("item not added when item id does not exist", func(t *testing.T) {
		mock.SetUp(t)

		itemId := uuid.New()

		tokenString := "token string"
		receiptString := "receipt string"

		verifierMock := mock.Mock[tokenVerifier]()
		storeMock := mock.Mock[foo]()

		mock.When(storeMock.GetItemById(itemId)).ThenReturn(nil, assert.AnError)

		jsonData := generateData(receiptString, itemId.String())
		ctx, recorder := setUpContext(tokenString, jsonData)
		handleBuyStoreItem(verifierMock, storeMock)(ctx)

		mock.Verify(storeMock, mock.Once()).GetItemById(itemId)
		mock.VerifyNoMoreInteractions(storeMock)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("item not added when receipt rarity and item rarity do not match", func(t *testing.T) {
		mock.SetUp(t)

		tokenString := "token string"
		receiptString := "receipt string"

		itemId := uuid.New()
		transactionId := uuid.New()

		item := models.Item{
			ItemId: itemId,
			Rarity: models.Uncommon,
		}

		receipt := &services.Receipt{
			Data: services.Data{
				TransactionId: transactionId,
				Product: services.Product{
					Rarity: models.Common,
				},
			},
		}

		verifierMock := mock.Mock[tokenVerifier]()
		storeMock := mock.Mock[foo]()

		mock.When(storeMock.GetItemById(itemId)).ThenReturn(item, nil)
		mock.When(verifierMock.VerifyReceipt(receiptString)).ThenReturn(receipt, nil)

		jsonData := generateData(receiptString, itemId.String())
		ctx, recorder := setUpContext(tokenString, jsonData)
		handleBuyStoreItem(verifierMock, storeMock)(ctx)

		mock.Verify(storeMock, mock.Once()).GetItemById(itemId)
		mock.VerifyNoMoreInteractions(storeMock)

		assert.Equal(t, http.StatusForbidden, recorder.Code)
	})

	t.Run("internal server error when add owned item fails", func(t *testing.T) {
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

		verifierMock := mock.Mock[tokenVerifier]()
		storeMock := mock.Mock[foo]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(token, nil)
		mock.When(verifierMock.VerifyReceipt(receiptString)).ThenReturn(receipt, nil)
		mock.When(storeMock.GetItemById(itemId)).ThenReturn(item, nil)
		mock.When(storeMock.AddOwnedItem(userId, itemId, transactionId)).ThenReturn(assert.AnError)

		jsonData := generateData(receiptString, itemId.String())
		ctx, recorder := setUpContext(tokenString, jsonData)
		handleBuyStoreItem(verifierMock, storeMock)(ctx)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
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

		verifierMock := mock.Mock[tokenVerifier]()
		storeMock := mock.Mock[foo]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(token, nil)
		mock.When(verifierMock.VerifyReceipt(receiptString)).ThenReturn(receipt, nil)
		mock.When(storeMock.GetItemById(itemId)).ThenReturn(item, nil)

		jsonData := generateData(receiptString, itemId.String())
		ctx, recorder := setUpContext(tokenString, jsonData)
		handleBuyStoreItem(verifierMock, storeMock)(ctx)

		mock.Verify(storeMock, mock.Once()).AddOwnedItem(userId, itemId, transactionId)
		assert.Equal(t, http.StatusNoContent, recorder.Code)
	})
}

func TestSetSelectedItem(t *testing.T) {

	generateData := func(itemId string) []byte {
		return []byte(fmt.Sprintf(`{
			"item_id": "%s"
		}`, itemId))
	}

	setUpContext := func(token string, jsonData []byte) (*gin.Context, *httptest.ResponseRecorder) {
		gin.SetMode(gin.TestMode)

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest("POST", "/items", bytes.NewBuffer(jsonData))

		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
		req.Header.Add("x-extension-jwt", token)

		ctx.Request = req
		return ctx, recorder
	}

	t.Run("pet not updated when extension token is invalid", func(t *testing.T) {
		mock.SetUp(t)

		tokenString := "token string"

		itemId := uuid.New()

		announcerMock := mock.Mock[updateAnnouncer]()
		verifierMock := mock.Mock[tokenVerifier]()
		storeMock := mock.Mock[bar]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(nil, services.NewErrInvalidToken(tokenString))

		jsonData := generateData(itemId.String())
		ctx, recorder := setUpContext(tokenString, jsonData)
		handleSetSelectedItem(announcerMock, verifierMock, storeMock)(ctx)

		mock.VerifyNoMoreInteractions(announcerMock)
		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run("pet not updated when extension token cannot be validated", func(t *testing.T) {
		mock.SetUp(t)

		tokenString := "token string"

		itemId := uuid.New()

		announcerMock := mock.Mock[updateAnnouncer]()
		verifierMock := mock.Mock[tokenVerifier]()
		storeMock := mock.Mock[bar]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(nil, assert.AnError)

		jsonData := generateData(itemId.String())
		ctx, recorder := setUpContext(tokenString, jsonData)
		handleSetSelectedItem(announcerMock, verifierMock, storeMock)(ctx)

		mock.VerifyNoMoreInteractions(announcerMock)
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("pet not updated when json has invalid format", func(t *testing.T) {
		mock.SetUp(t)

		tokenString := "token string"

		announcerMock := mock.Mock[updateAnnouncer]()
		verifierMock := mock.Mock[tokenVerifier]()
		storeMock := mock.Mock[bar]()

		jsonData := make([]byte, 0)
		ctx, recorder := setUpContext(tokenString, jsonData)
		handleSetSelectedItem(announcerMock, verifierMock, storeMock)(ctx)

		mock.VerifyNoMoreInteractions(announcerMock)
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("pet not updated when item id is not a valid uuid", func(t *testing.T) {
		mock.SetUp(t)

		tokenString := "token string"
		itemId := "invalid id"

		announcerMock := mock.Mock[updateAnnouncer]()
		verifierMock := mock.Mock[tokenVerifier]()
		storeMock := mock.Mock[bar]()

		jsonData := generateData(itemId)
		ctx, recorder := setUpContext(tokenString, jsonData)
		handleSetSelectedItem(announcerMock, verifierMock, storeMock)(ctx)

		mock.VerifyNoMoreInteractions(announcerMock)
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("pet not updated when item id does not exist", func(t *testing.T) {
		mock.SetUp(t)

		tokenString := "token string"
		itemId := uuid.New()

		announcerMock := mock.Mock[updateAnnouncer]()
		verifierMock := mock.Mock[tokenVerifier]()
		storeMock := mock.Mock[bar]()

		mock.When(storeMock.GetItemById(itemId)).ThenReturn(nil, assert.AnError)

		jsonData := generateData(itemId.String())
		ctx, recorder := setUpContext(tokenString, jsonData)
		handleSetSelectedItem(announcerMock, verifierMock, storeMock)(ctx)

		mock.VerifyNoMoreInteractions(announcerMock)
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("pet not updated when item unowned", func(t *testing.T) {
		mock.SetUp(t)

		tokenString := "token string"
		channelId := twitch.Id("channel id")
		userId := twitch.Id("user id")
		itemId := uuid.New()

		token := &services.ExtToken{
			ChannelId: channelId,
			UserId:    userId,
		}

		announcerMock := mock.Mock[updateAnnouncer]()
		verifierMock := mock.Mock[tokenVerifier]()
		storeMock := mock.Mock[bar]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(token, nil)
		mock.When(storeMock.SetSelectedItem(userId, channelId, itemId)).ThenReturn(services.ErrSelectUnownedItem)

		jsonData := generateData(itemId.String())
		ctx, recorder := setUpContext(tokenString, jsonData)
		handleSetSelectedItem(announcerMock, verifierMock, storeMock)(ctx)

		mock.VerifyNoMoreInteractions(announcerMock)
		assert.Equal(t, http.StatusForbidden, recorder.Code)
	})

	t.Run("pet not updated when item unowned", func(t *testing.T) {
		mock.SetUp(t)

		tokenString := "token string"
		channelId := twitch.Id("channel id")
		userId := twitch.Id("user id")
		itemId := uuid.New()

		token := &services.ExtToken{
			ChannelId: channelId,
			UserId:    userId,
		}

		announcerMock := mock.Mock[updateAnnouncer]()
		verifierMock := mock.Mock[tokenVerifier]()
		storeMock := mock.Mock[bar]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(token, nil)
		mock.When(storeMock.SetSelectedItem(userId, channelId, itemId)).ThenReturn(assert.AnError)

		jsonData := generateData(itemId.String())
		ctx, recorder := setUpContext(tokenString, jsonData)
		handleSetSelectedItem(announcerMock, verifierMock, storeMock)(ctx)

		mock.VerifyNoMoreInteractions(announcerMock)
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
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

		announcerMock := mock.Mock[updateAnnouncer]()
		verifierMock := mock.Mock[tokenVerifier]()
		storeMock := mock.Mock[bar]()

		mock.When(verifierMock.VerifyExtToken(tokenString)).ThenReturn(&token, nil)
		mock.When(storeMock.GetItemById(itemId)).ThenReturn(item, nil)

		jsonData := generateData(itemId.String())
		ctx, recorder := setUpContext(tokenString, jsonData)
		handleSetSelectedItem(announcerMock, verifierMock, storeMock)(ctx)

		mock.Verify(storeMock, mock.Once()).SetSelectedItem(userId, channelId, itemId)
		mock.Verify(announcerMock, mock.Once()).AnnounceUpdate(channelId, userId, image)

		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

func TestAddUserToChannel(t *testing.T) {

	generateData := func(userId twitch.Id, username string) []byte {
		return []byte(fmt.Sprintf(`{
			"user_id": "%s",
			"username": "%s"
			}`, userId, username))
	}

	setUpContext := func(channelId twitch.Id, jsonData []byte) (*gin.Context, *httptest.ResponseRecorder) {
		gin.SetMode(gin.TestMode)

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest("", "", bytes.NewBuffer(jsonData))
		ctx.Params = gin.Params{{Key: ChannelId, Value: string(channelId)}}

		ctx.Request = req
		return ctx, recorder
	}

	t.Run("bad request when json has invalid format", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.Id("channel id")

		announcerMock := mock.Mock[joinAnnouncer]()
		petsMock := mock.Mock[petGetter]()

		jsonData := make([]byte, 0)
		ctx, recorder := setUpContext(channelId, jsonData)
		handleAddPetToChannel(announcerMock, petsMock)(ctx)

		mock.VerifyNoMoreInteractions(announcerMock)
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("internal server error when get pet fails", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.Id("channel id")
		userId := twitch.Id("user id")
		username := "username"

		announcerMock := mock.Mock[joinAnnouncer]()
		petsMock := mock.Mock[petGetter]()

		mock.When(petsMock.GetPet(userId, channelId, username)).ThenReturn(nil, assert.AnError)

		jsonData := generateData(userId, username)
		ctx, recorder := setUpContext(channelId, jsonData)
		handleAddPetToChannel(announcerMock, petsMock)(ctx)

		mock.VerifyNoMoreInteractions(announcerMock)
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("join announced and status no content", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.Id("channel id")
		userId := twitch.Id("user id")
		username := "username"

		pet := services.Pet{Username: username}

		announcerMock := mock.Mock[joinAnnouncer]()
		petsMock := mock.Mock[petGetter]()

		mock.When(petsMock.GetPet(userId, channelId, username)).ThenReturn(pet, nil)

		jsonData := generateData(userId, username)
		ctx, recorder := setUpContext(channelId, jsonData)
		handleAddPetToChannel(announcerMock, petsMock)(ctx)

		mock.Verify(announcerMock, mock.Once()).AnnounceJoin(channelId, pet)
		assert.Equal(t, http.StatusNoContent, recorder.Code)
	})
}

func TestRemoveUserFromChannel(t *testing.T) {

	setUpContext := func(channelId, userId twitch.Id) (*gin.Context, *httptest.ResponseRecorder) {
		gin.SetMode(gin.TestMode)

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Params = gin.Params{
			{Key: ChannelId, Value: string(channelId)},
			{Key: UserId, Value: string(userId)},
		}

		return ctx, recorder
	}

	mock.SetUp(t)

	channelId := twitch.Id("channel id")
	userId := twitch.Id("user id")

	announcerMock := mock.Mock[partAnnouncer]()

	ctx, recorder := setUpContext(channelId, userId)
	handleRemoveUserFromChannel(announcerMock)(ctx)

	mock.Verify(announcerMock, mock.Once()).AnnouncePart(channelId, userId)
	assert.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestAction(t *testing.T) {
	mock.SetUp(t)

	setUpContext := func(channelId, userId twitch.Id, action string) (*gin.Context, *httptest.ResponseRecorder) {
		gin.SetMode(gin.TestMode)

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Params = gin.Params{
			{Key: ChannelId, Value: string(channelId)},
			{Key: UserId, Value: string(userId)},
			{Key: Action, Value: action},
		}

		return ctx, recorder
	}

	channelId := twitch.Id("channel id")
	userId := twitch.Id("user id")
	action := "action"

	announcerMock := mock.Mock[actionAnnouncer]()

	ctx, recorder := setUpContext(channelId, userId, action)
	handleAction(announcerMock)(ctx)

	mock.Verify(announcerMock, mock.Once()).AnnounceAction(channelId, userId, action)
	assert.Equal(t, http.StatusNoContent, recorder.Code)
}
