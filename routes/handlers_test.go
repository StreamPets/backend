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
	"github.com/streampets/backend/database"
	"github.com/streampets/backend/items"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/pets"
	"github.com/streampets/backend/test"
	"github.com/streampets/backend/twitch"
	"github.com/stretchr/testify/assert"
)

func TestHandleLogin(t *testing.T) {
	setUpContext := func(channelId twitch.UserId) (*gin.Context, *httptest.ResponseRecorder) {
		gin.SetMode(gin.TestMode)

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest("GET", "/items", nil)
		ctx.Set(ChannelId, string(channelId))

		ctx.Request = req
		return ctx, recorder
	}

	type mockDep interface {
		GetOverlayId(channelId twitch.UserId) (uuid.UUID, error)
	}

	t.Run("status bad request when channel id has no overlay id", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.UserId("channel id")
		ctx, recorder := setUpContext(channelId)

		mockDep := mock.Mock[mockDep]()
		mock.When(mockDep.GetOverlayId(channelId)).ThenReturn(nil, database.ErrNoOverlayId)

		handleLogin(mockDep.GetOverlayId)(ctx)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("internal server error when get overlay id fails", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.UserId("channel id")
		ctx, recorder := setUpContext(channelId)

		mockDep := mock.Mock[mockDep]()
		mock.When(mockDep.GetOverlayId(channelId)).ThenReturn(nil, assert.AnError)

		handleLogin(mockDep.GetOverlayId)(ctx)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("channel id and overlay id returned in normal case", func(t *testing.T) {
		mock.SetUp(t)

		type userData struct {
			OverlayId uuid.UUID     `json:"overlay_id"`
			ChannelId twitch.UserId `json:"channel_id"`
		}

		channelId := twitch.UserId("channel id")
		overlayId := uuid.New()

		ctx, recorder := setUpContext(channelId)

		mockDep := mock.Mock[mockDep]()
		mock.When(mockDep.GetOverlayId(channelId)).ThenReturn(overlayId, nil)

		handleLogin(mockDep.GetOverlayId)(ctx)

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

	setUpContext := func(channelId twitch.UserId) (*gin.Context, *test.CloseNotifierResponseWriter) {
		gin.SetMode(gin.TestMode)

		recorder := &test.CloseNotifierResponseWriter{ResponseRecorder: httptest.NewRecorder()}
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Set(ChannelId, string(channelId))

		return ctx, recorder
	}

	type mockDep interface {
		AddClient(channelId twitch.UserId) announcers.Client
		RemoveClient(client announcers.Client)
	}

	t.Run("receive and send events from stream", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.UserId("channel id")
		ctx, recorder := setUpContext(channelId)

		stream := make(chan announcers.Announcement)
		client := announcers.Client{Stream: stream}

		mockDep := mock.Mock[mockDep]()

		mock.When(mockDep.AddClient(channelId)).ThenReturn(client)

		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			handleListen(
				mockDep.AddClient,
				mockDep.RemoveClient,
			)(ctx)
		}()

		stream <- announcers.Announcement{
			Event:   "event",
			Message: "message",
		}

		close(stream)
		wg.Wait()

		mock.Verify(mockDep, mock.Once()).AddClient(channelId)
		mock.Verify(mockDep, mock.Once()).RemoveClient(client)
		mock.VerifyNoMoreInteractions(mockDep)

		assert.Contains(t, recorder.Body.String(), "event:event")
		assert.Contains(t, recorder.Body.String(), "data:message")
	})
}

func TestGetStoreData(t *testing.T) {

	setUpContext := func(channelId, userId twitch.UserId) (*gin.Context, *httptest.ResponseRecorder) {
		gin.SetMode(gin.TestMode)

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest("GET", "/items", nil)

		ctx.Set(ChannelId, string(channelId))
		ctx.Set(UserId, string(userId))

		ctx.Request = req
		return ctx, recorder
	}

	type mockDep interface {
		GetChannelsItems(channelId twitch.UserId) ([]models.Item, error)
	}

	t.Run("internal server error when error received from get channels items", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.UserId("channel id")
		userId := twitch.UserId("user id")

		mockDep := mock.Mock[mockDep]()

		mock.When(mockDep.GetChannelsItems(channelId)).ThenReturn(nil, assert.AnError)

		ctx, recorder := setUpContext(channelId, userId)
		handleGetStoreData(
			mockDep.GetChannelsItems,
		)(ctx)

		mock.Verify(mockDep, mock.Once()).GetChannelsItems(channelId)
		mock.VerifyNoMoreInteractions(mockDep)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("items returned when extension token and channel id are valid", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.UserId("channel id")
		userId := twitch.UserId("user id")

		storeItems := []models.Item{{}, {}}

		mockDep := mock.Mock[mockDep]()

		mock.When(mockDep.GetChannelsItems(channelId)).ThenReturn(storeItems, nil)

		ctx, recorder := setUpContext(channelId, userId)
		handleGetStoreData(
			mockDep.GetChannelsItems,
		)(ctx)

		mock.Verify(mockDep, mock.Once()).GetChannelsItems(channelId)
		mock.VerifyNoMoreInteractions(mockDep)

		assert.Equal(t, recorder.Code, http.StatusOK)

		var response []models.Item
		if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
			t.Errorf("could not parse json response")
		}

		assert.Equal(t, storeItems, response)
	})
}

func TestGetUserData(t *testing.T) {

	setUpContext := func(channelId, userId twitch.UserId) (*gin.Context, *httptest.ResponseRecorder) {
		gin.SetMode(gin.TestMode)

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest("GET", "/items", nil)

		ctx.Set(ChannelId, string(channelId))
		ctx.Set(UserId, string(userId))

		ctx.Request = req
		return ctx, recorder
	}

	type mockDep interface {
		GetSelectedItem(userId, channelId twitch.UserId) (models.Item, error)
		GetOwnedItems(channelId, userId twitch.UserId) ([]models.Item, error)
	}

	t.Run("internal server error when get owned items fails", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.UserId("channel id")
		userId := twitch.UserId("user id")

		mockDep := mock.Mock[mockDep]()

		mock.When(mockDep.GetOwnedItems(channelId, userId)).ThenReturn(nil, assert.AnError)

		ctx, recorder := setUpContext(channelId, userId)
		handleGetUserData(
			mockDep.GetSelectedItem,
			mockDep.GetOwnedItems,
		)(ctx)

		mock.Verify(mockDep, mock.Once()).GetOwnedItems(channelId, userId)
		mock.VerifyNoMoreInteractions(mockDep)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("internal server error when get selected item fails", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.UserId("channel id")
		userId := twitch.UserId("user id")

		mockDep := mock.Mock[mockDep]()

		mock.When(mockDep.GetSelectedItem(userId, channelId)).ThenReturn(nil, assert.AnError)

		ctx, recorder := setUpContext(channelId, userId)
		handleGetUserData(
			mockDep.GetSelectedItem,
			mockDep.GetOwnedItems,
		)(ctx)

		mock.Verify(mockDep, mock.Once()).GetOwnedItems(channelId, userId)
		mock.Verify(mockDep, mock.Once()).GetSelectedItem(userId, channelId)
		mock.VerifyNoMoreInteractions(mockDep)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("status ok when all pre-requisites are met", func(t *testing.T) {
		mock.SetUp(t)

		type Response struct {
			OwnedItems   []models.Item `json:"owned"`
			SelectedItem models.Item   `json:"selected"`
		}

		channelId := twitch.UserId("channel id")
		userId := twitch.UserId("user id")

		selectedItem := models.Item{ItemId: uuid.New()}
		ownedItems := []models.Item{selectedItem}

		mockDep := mock.Mock[mockDep]()

		mock.When(mockDep.GetOwnedItems(channelId, userId)).ThenReturn(ownedItems, nil)
		mock.When(mockDep.GetSelectedItem(userId, channelId)).ThenReturn(selectedItem, nil)

		ctx, recorder := setUpContext(channelId, userId)
		handleGetUserData(
			mockDep.GetSelectedItem,
			mockDep.GetOwnedItems,
		)(ctx)

		mock.Verify(mockDep, mock.Once()).GetOwnedItems(channelId, userId)
		mock.Verify(mockDep, mock.Once()).GetSelectedItem(userId, channelId)
		mock.VerifyNoMoreInteractions(mockDep)

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

	setUpContext := func(userId twitch.UserId, itemId, transactionId uuid.UUID, rarity models.Rarity) (*gin.Context, *httptest.ResponseRecorder) {
		gin.SetMode(gin.TestMode)

		jsonData := []byte(fmt.Sprintf(`{
			"item_id": "%s"
		}`, itemId))

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest("POST", "/items", bytes.NewBuffer(jsonData))

		ctx.Set(UserId, string(userId))
		ctx.Set(TransactionId, transactionId.String())
		ctx.Set(Rarity, string(rarity))

		ctx.Request = req
		return ctx, recorder
	}

	type mockDep interface {
		GetItemById(itemId uuid.UUID) (models.Item, error)
		AddOwnedItem(userId twitch.UserId, itemId, transactionId uuid.UUID) error
	}

	t.Run("item not added when item with item id does not exist", func(t *testing.T) {
		mock.SetUp(t)

		userId := twitch.UserId("user id")
		itemId := uuid.New()
		transactionId := uuid.New()
		rarity := models.Common

		mockDep := mock.Mock[mockDep]()

		err := database.ErrItemNotFoundById
		mock.When(mockDep.GetItemById(itemId)).ThenReturn(nil, err)

		ctx, recorder := setUpContext(userId, itemId, transactionId, rarity)
		handleBuyStoreItem(
			mockDep.GetItemById,
			mockDep.AddOwnedItem,
		)(ctx)

		mock.Verify(mockDep, mock.Once()).GetItemById(itemId)
		mock.VerifyNoMoreInteractions(mockDep)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("item not added when get item by id fails", func(t *testing.T) {
		mock.SetUp(t)

		userId := twitch.UserId("user id")
		itemId := uuid.New()
		transactionId := uuid.New()
		rarity := models.Common

		mockDep := mock.Mock[mockDep]()

		mock.When(mockDep.GetItemById(itemId)).ThenReturn(nil, assert.AnError)

		ctx, recorder := setUpContext(userId, itemId, transactionId, rarity)
		handleBuyStoreItem(
			mockDep.GetItemById,
			mockDep.AddOwnedItem,
		)(ctx)

		mock.Verify(mockDep, mock.Once()).GetItemById(itemId)
		mock.VerifyNoMoreInteractions(mockDep)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("item not added when receipt rarity and item rarity do not match", func(t *testing.T) {
		mock.SetUp(t)

		userId := twitch.UserId("user id")
		itemId := uuid.New()
		transactionId := uuid.New()
		rarity := models.Common

		item := models.Item{
			ItemId: itemId,
			Rarity: models.Uncommon,
		}

		mockDep := mock.Mock[mockDep]()
		mock.When(mockDep.GetItemById(itemId)).ThenReturn(item, nil)

		ctx, recorder := setUpContext(userId, itemId, transactionId, rarity)
		handleBuyStoreItem(
			mockDep.GetItemById,
			mockDep.AddOwnedItem,
		)(ctx)

		mock.Verify(mockDep, mock.Once()).GetItemById(itemId)
		mock.VerifyNoMoreInteractions(mockDep)

		assert.Equal(t, http.StatusForbidden, recorder.Code)
	})

	t.Run("internal server error when add owned item fails", func(t *testing.T) {
		mock.SetUp(t)

		userId := twitch.UserId("user id")
		itemId := uuid.New()
		transactionId := uuid.New()
		rarity := models.Common

		item := models.Item{
			ItemId: itemId,
			Rarity: models.Common,
		}

		mockDep := mock.Mock[mockDep]()

		mock.When(mockDep.GetItemById(itemId)).ThenReturn(item, nil)
		mock.When(mockDep.AddOwnedItem(userId, itemId, transactionId)).ThenReturn(assert.AnError)

		ctx, recorder := setUpContext(userId, itemId, transactionId, rarity)
		handleBuyStoreItem(
			mockDep.GetItemById,
			mockDep.AddOwnedItem,
		)(ctx)

		mock.Verify(mockDep, mock.Once()).GetItemById(itemId)
		mock.Verify(mockDep, mock.Once()).AddOwnedItem(userId, itemId, transactionId)
		mock.VerifyNoMoreInteractions(mockDep)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("item added when all pre-requisites are met", func(t *testing.T) {
		mock.SetUp(t)

		userId := twitch.UserId("user id")
		itemId := uuid.New()
		transactionId := uuid.New()
		rarity := models.Common

		item := models.Item{
			ItemId: itemId,
			Rarity: models.Common,
		}

		mockDep := mock.Mock[mockDep]()

		mock.When(mockDep.GetItemById(itemId)).ThenReturn(item, nil)

		ctx, recorder := setUpContext(userId, itemId, transactionId, rarity)
		handleBuyStoreItem(
			mockDep.GetItemById,
			mockDep.AddOwnedItem,
		)(ctx)

		mock.Verify(mockDep, mock.Once()).GetItemById(itemId)
		mock.Verify(mockDep, mock.Once()).AddOwnedItem(userId, itemId, transactionId)
		mock.VerifyNoMoreInteractions(mockDep)

		assert.Equal(t, http.StatusNoContent, recorder.Code)
	})
}

func TestSetSelectedItem(t *testing.T) {

	generateData := func(itemId string) []byte {
		return []byte(fmt.Sprintf(`{
			"item_id": "%s"
		}`, itemId))
	}

	setUpContext := func(channelId, userId twitch.UserId, jsonData []byte) (*gin.Context, *httptest.ResponseRecorder) {
		gin.SetMode(gin.TestMode)

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest("POST", "/items", bytes.NewBuffer(jsonData))

		ctx.Set(ChannelId, string(channelId))
		ctx.Set(UserId, string(userId))

		ctx.Request = req
		return ctx, recorder
	}

	type mockDep interface {
		AnnounceUpdate(channelId, userId twitch.UserId, image string)
		GetItemById(itemId uuid.UUID) (models.Item, error)
		SetSelectedItem(userId, channelId twitch.UserId, itemId uuid.UUID) error
	}

	t.Run("pet not updated when json has invalid format", func(t *testing.T) {
		mock.SetUp(t)

		mockDep := mock.Mock[mockDep]()

		jsonData := make([]byte, 0)
		ctx, recorder := setUpContext("", "", jsonData)
		handleSetSelectedItem(
			mockDep.AnnounceUpdate,
			mockDep.GetItemById,
			mockDep.SetSelectedItem,
		)(ctx)

		mock.VerifyNoMoreInteractions(mockDep)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("pet not updated when item id is not a valid uuid", func(t *testing.T) {
		mock.SetUp(t)

		itemId := "invalid id"

		mockDep := mock.Mock[mockDep]()

		jsonData := generateData(itemId)
		ctx, recorder := setUpContext("", "", jsonData)
		handleSetSelectedItem(
			mockDep.AnnounceUpdate,
			mockDep.GetItemById,
			mockDep.SetSelectedItem,
		)(ctx)

		mock.VerifyNoMoreInteractions(mockDep)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("pet not updated when item id does not exist", func(t *testing.T) {
		mock.SetUp(t)

		itemId := uuid.New()

		mockDep := mock.Mock[mockDep]()

		err := database.ErrItemNotFoundById
		mock.When(mockDep.GetItemById(itemId)).ThenReturn(nil, err)

		jsonData := generateData(itemId.String())
		ctx, recorder := setUpContext("", "", jsonData)
		handleSetSelectedItem(
			mockDep.AnnounceUpdate,
			mockDep.GetItemById,
			mockDep.SetSelectedItem,
		)(ctx)

		mock.Verify(mockDep, mock.Once()).GetItemById(itemId)
		mock.VerifyNoMoreInteractions(mockDep)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("pet not updated when get item by id fails", func(t *testing.T) {
		mock.SetUp(t)

		itemId := uuid.New()

		mockDep := mock.Mock[mockDep]()

		mock.When(mockDep.GetItemById(itemId)).ThenReturn(nil, assert.AnError)

		jsonData := generateData(itemId.String())
		ctx, recorder := setUpContext("", "", jsonData)
		handleSetSelectedItem(
			mockDep.AnnounceUpdate,
			mockDep.GetItemById,
			mockDep.SetSelectedItem,
		)(ctx)

		mock.Verify(mockDep, mock.Once()).GetItemById(itemId)
		mock.VerifyNoMoreInteractions(mockDep)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("pet not updated when item unowned", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.UserId("channel id")
		userId := twitch.UserId("user id")
		itemId := uuid.New()

		mockDep := mock.Mock[mockDep]()

		err := items.ErrSelectUnownedItem
		mock.When(mockDep.SetSelectedItem(userId, channelId, itemId)).ThenReturn(err)

		jsonData := generateData(itemId.String())
		ctx, recorder := setUpContext(channelId, userId, jsonData)
		handleSetSelectedItem(
			mockDep.AnnounceUpdate,
			mockDep.GetItemById,
			mockDep.SetSelectedItem,
		)(ctx)

		mock.Verify(mockDep, mock.Once()).GetItemById(itemId)
		mock.Verify(mockDep, mock.Once()).SetSelectedItem(userId, channelId, itemId)
		mock.VerifyNoMoreInteractions(mockDep)

		assert.Equal(t, http.StatusForbidden, recorder.Code)
	})

	t.Run("pet not updated when item unowned", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.UserId("channel id")
		userId := twitch.UserId("user id")
		itemId := uuid.New()

		mockDep := mock.Mock[mockDep]()

		mock.When(mockDep.SetSelectedItem(userId, channelId, itemId)).ThenReturn(assert.AnError)

		jsonData := generateData(itemId.String())
		ctx, recorder := setUpContext(channelId, userId, jsonData)
		handleSetSelectedItem(
			mockDep.AnnounceUpdate,
			mockDep.GetItemById,
			mockDep.SetSelectedItem,
		)(ctx)

		mock.Verify(mockDep, mock.Once()).GetItemById(itemId)
		mock.Verify(mockDep, mock.Once()).SetSelectedItem(userId, channelId, itemId)
		mock.VerifyNoMoreInteractions(mockDep)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("pet updated when pre-requisites are met", func(t *testing.T) {
		mock.SetUp(t)

		image := "image"

		channelId := twitch.UserId("channel id")
		userId := twitch.UserId("user id")
		itemId := uuid.New()

		item := models.Item{ItemId: itemId, Image: image}

		mockDep := mock.Mock[mockDep]()

		mock.When(mockDep.GetItemById(itemId)).ThenReturn(item, nil)

		jsonData := generateData(itemId.String())
		ctx, recorder := setUpContext(channelId, userId, jsonData)
		handleSetSelectedItem(
			mockDep.AnnounceUpdate,
			mockDep.GetItemById,
			mockDep.SetSelectedItem,
		)(ctx)

		mock.Verify(mockDep, mock.Once()).GetItemById(itemId)
		mock.Verify(mockDep, mock.Once()).SetSelectedItem(userId, channelId, itemId)
		mock.Verify(mockDep, mock.Once()).AnnounceUpdate(channelId, userId, image)
		mock.VerifyNoMoreInteractions(mockDep)

		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

func TestAddUserToChannel(t *testing.T) {

	generateData := func(userId twitch.UserId, username string) []byte {
		return []byte(fmt.Sprintf(`{
			"user_id": "%s",
			"username": "%s"
			}`, userId, username))
	}

	setUpContext := func(channelId twitch.UserId, jsonData []byte) (*gin.Context, *httptest.ResponseRecorder) {
		gin.SetMode(gin.TestMode)

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest("", "", bytes.NewBuffer(jsonData))
		ctx.Params = gin.Params{{Key: ChannelId, Value: string(channelId)}}

		ctx.Request = req
		return ctx, recorder
	}

	type mockDep interface {
		AnnounceJoin(channelId twitch.UserId, pet pets.Pet)
		GetPet(userId, channelId twitch.UserId, username string) (pets.Pet, error)
	}

	t.Run("bad request when json has invalid format", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.UserId("channel id")

		mockDep := mock.Mock[mockDep]()

		jsonData := make([]byte, 0)
		ctx, recorder := setUpContext(channelId, jsonData)
		handleAddPetToChannel(
			mockDep.AnnounceJoin,
			mockDep.GetPet,
		)(ctx)

		mock.VerifyNoMoreInteractions(mockDep)
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("internal server error when get pet fails", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.UserId("channel id")
		userId := twitch.UserId("user id")
		username := "username"

		mockDep := mock.Mock[mockDep]()

		mock.When(mockDep.GetPet(userId, channelId, username)).ThenReturn(nil, assert.AnError)

		jsonData := generateData(userId, username)
		ctx, recorder := setUpContext(channelId, jsonData)
		handleAddPetToChannel(
			mockDep.AnnounceJoin,
			mockDep.GetPet,
		)(ctx)

		mock.Verify(mockDep, mock.Once()).GetPet(userId, channelId, username)
		mock.VerifyNoMoreInteractions(mockDep)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("join announced and status no content", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.UserId("channel id")
		userId := twitch.UserId("user id")
		username := "username"

		pet := pets.Pet{Username: username}

		mockDep := mock.Mock[mockDep]()

		mock.When(mockDep.GetPet(userId, channelId, username)).ThenReturn(pet, nil)

		jsonData := generateData(userId, username)
		ctx, recorder := setUpContext(channelId, jsonData)
		handleAddPetToChannel(
			mockDep.AnnounceJoin,
			mockDep.GetPet,
		)(ctx)

		mock.Verify(mockDep, mock.Once()).GetPet(userId, channelId, username)
		mock.Verify(mockDep, mock.Once()).AnnounceJoin(channelId, pet)
		mock.VerifyNoMoreInteractions(mockDep)

		assert.Equal(t, http.StatusNoContent, recorder.Code)
	})
}

func TestRemoveUserFromChannel(t *testing.T) {

	setUpContext := func(channelId, userId twitch.UserId) (*gin.Context, *httptest.ResponseRecorder) {
		gin.SetMode(gin.TestMode)

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Params = gin.Params{
			{Key: ChannelId, Value: string(channelId)},
			{Key: UserId, Value: string(userId)},
		}

		return ctx, recorder
	}

	type mockDep interface {
		AnnouncePart(channelId, userId twitch.UserId)
	}

	mock.SetUp(t)

	channelId := twitch.UserId("channel id")
	userId := twitch.UserId("user id")

	dep := mock.Mock[mockDep]()

	ctx, recorder := setUpContext(channelId, userId)
	handleRemoveUserFromChannel(
		dep.AnnouncePart,
	)(ctx)

	mock.Verify(dep, mock.Once()).AnnouncePart(channelId, userId)
	mock.VerifyNoMoreInteractions(dep)

	assert.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestAction(t *testing.T) {

	setUpContext := func(channelId, userId twitch.UserId, action string) (*gin.Context, *httptest.ResponseRecorder) {
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

	type mockDep interface {
		AnnounceAction(channelId, userId twitch.UserId, action string)
	}

	mock.SetUp(t)

	channelId := twitch.UserId("channel id")
	userId := twitch.UserId("user id")
	action := "action"

	dep := mock.Mock[mockDep]()

	ctx, recorder := setUpContext(channelId, userId, action)
	handleAction(
		dep.AnnounceAction,
	)(ctx)

	mock.Verify(dep, mock.Once()).AnnounceAction(channelId, userId, action)
	mock.VerifyNoMoreInteractions(dep)

	assert.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestUpdateUser(t *testing.T) {

	generateData := func(itemName string) []byte {
		return []byte(fmt.Sprintf(`{
			"item_name": "%s"
		}`, itemName))
	}

	setUpContext := func(channelId, userId twitch.UserId, jsonData []byte) (*gin.Context, *httptest.ResponseRecorder) {
		gin.SetMode(gin.TestMode)

		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		req, _ := http.NewRequest("PUT", "", bytes.NewBuffer(jsonData))
		ctx.Params = gin.Params{
			{Key: ChannelId, Value: string(channelId)},
			{Key: UserId, Value: string(userId)},
		}
		ctx.Request = req

		return ctx, recorder
	}

	type mockDep interface {
		AnnounceUpdate(channelId, userId twitch.UserId, image string)
		GetItemByName(channelId twitch.UserId, itemName string) (models.Item, error)
		SetSelectedItem(userId, channelId twitch.UserId, itemId uuid.UUID) error
	}

	t.Run("bad request when item not found", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.UserId("channel id")
		userId := twitch.UserId("user id")
		itemName := "item name"

		mockDep := mock.Mock[mockDep]()

		err := database.ErrItemNotFoundByName
		mock.When(mockDep.GetItemByName(channelId, itemName)).ThenReturn(nil, err)

		jsonData := generateData(itemName)
		ctx, recorder := setUpContext(channelId, userId, jsonData)
		handleUpdate(
			mockDep.AnnounceUpdate,
			mockDep.GetItemByName,
			mockDep.SetSelectedItem,
		)(ctx)

		mock.Verify(mockDep, mock.Once()).GetItemByName(channelId, itemName)
		mock.VerifyNoMoreInteractions(mockDep)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("status forbidden when item not owned", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.UserId("channel id")
		userId := twitch.UserId("user id")
		itemName := "item name"

		itemId := uuid.New()
		image := "image"

		item := models.Item{
			ItemId: itemId,
			Image:  image,
		}

		mockDep := mock.Mock[mockDep]()

		err := items.ErrSelectUnownedItem
		mock.When(mockDep.GetItemByName(channelId, itemName)).ThenReturn(item, nil)
		mock.When(mockDep.SetSelectedItem(userId, channelId, itemId)).ThenReturn(err)

		jsonData := generateData(itemName)
		ctx, recorder := setUpContext(channelId, userId, jsonData)
		handleUpdate(
			mockDep.AnnounceUpdate,
			mockDep.GetItemByName,
			mockDep.SetSelectedItem,
		)(ctx)

		mock.Verify(mockDep, mock.Once()).GetItemByName(channelId, itemName)
		mock.Verify(mockDep, mock.Once()).SetSelectedItem(userId, channelId, itemId)
		mock.VerifyNoMoreInteractions(mockDep)

		assert.Equal(t, http.StatusForbidden, recorder.Code)
	})

	t.Run("internal server error when set selected item fails", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.UserId("channel id")
		userId := twitch.UserId("user id")
		itemName := "item name"

		itemId := uuid.New()
		image := "image"

		item := models.Item{
			ItemId: itemId,
			Image:  image,
		}

		mockDep := mock.Mock[mockDep]()

		mock.When(mockDep.GetItemByName(channelId, itemName)).ThenReturn(item, nil)
		mock.When(mockDep.SetSelectedItem(userId, channelId, itemId)).ThenReturn(assert.AnError)

		jsonData := generateData(itemName)
		ctx, recorder := setUpContext(channelId, userId, jsonData)
		handleUpdate(
			mockDep.AnnounceUpdate,
			mockDep.GetItemByName,
			mockDep.SetSelectedItem,
		)(ctx)

		mock.Verify(mockDep, mock.Once()).GetItemByName(channelId, itemName)
		mock.Verify(mockDep, mock.Once()).SetSelectedItem(userId, channelId, itemId)
		mock.VerifyNoMoreInteractions(mockDep)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("status no content for regular case", func(t *testing.T) {
		mock.SetUp(t)

		channelId := twitch.UserId("channel id")
		userId := twitch.UserId("user id")
		itemName := "item name"

		itemId := uuid.New()
		image := "image"

		item := models.Item{
			ItemId: itemId,
			Image:  image,
		}

		mockDep := mock.Mock[mockDep]()

		mock.When(mockDep.GetItemByName(channelId, itemName)).ThenReturn(item, nil)

		jsonData := generateData(itemName)
		ctx, recorder := setUpContext(channelId, userId, jsonData)
		handleUpdate(
			mockDep.AnnounceUpdate,
			mockDep.GetItemByName,
			mockDep.SetSelectedItem,
		)(ctx)

		mock.Verify(mockDep, mock.Once()).GetItemByName(channelId, itemName)
		mock.Verify(mockDep, mock.Once()).SetSelectedItem(userId, channelId, itemId)
		mock.Verify(mockDep, mock.Once()).AnnounceUpdate(channelId, userId, image)
		mock.VerifyNoMoreInteractions(mockDep)

		assert.Equal(t, http.StatusNoContent, recorder.Code)
	})
}
