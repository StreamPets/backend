package repositories

import (
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/test"
)

func TestGetSelectedItem(t *testing.T) {
	channelId := models.TwitchId("channel id")
	userId := models.TwitchId("user id")

	itemId := uuid.New()
	item := models.Item{ItemId: itemId}

	selectedItem := models.SelectedItem{
		UserId:    userId,
		ChannelId: channelId,
		ItemId:    itemId,
	}

	db := test.CreateTestDB()
	if result := db.Create(&item); result.Error != nil {
		panic(result.Error)
	}
	if result := db.Create(&selectedItem); result.Error != nil {
		panic(result.Error)
	}

	itemRepo := NewItemRepository(db)
	got, err := itemRepo.GetSelectedItem(userId, channelId)

	assertNoError(err, t)
	assertItemsEqual(got, item, t)
}

func TestSetSelectedItem(t *testing.T) {
	channelId := models.TwitchId("channel id")
	userId := models.TwitchId("user id")

	itemId := uuid.New()
	item := models.Item{ItemId: itemId}

	newItemId := uuid.New()
	newItem := models.Item{ItemId: newItemId}

	selectedItem := models.SelectedItem{
		UserId:    userId,
		ChannelId: channelId,
		ItemId:    itemId,
	}

	db := test.CreateTestDB()
	if result := db.Create(&item); result.Error != nil {
		panic(result.Error)
	}
	if result := db.Create(&newItem); result.Error != nil {
		panic(result.Error)
	}
	if result := db.Create(&selectedItem); result.Error != nil {
		panic(result.Error)
	}

	itemRepo := NewItemRepository(db)

	got, err := itemRepo.GetSelectedItem(userId, channelId)
	assertNoError(err, t)
	assertItemsEqual(got, item, t)

	err = itemRepo.SetSelectedItem(userId, channelId, newItemId)
	assertNoError(err, t)

	got, err = itemRepo.GetSelectedItem(userId, channelId)
	assertNoError(err, t)
	assertItemsEqual(got, newItem, t)
}

func TestDeleteSelectedItem(t *testing.T) {
	userId := models.TwitchId("user id")
	channelId := models.TwitchId("twitch id")

	itemId := uuid.New()

	selectedItem := models.SelectedItem{
		ItemId: itemId,
	}

	db := test.CreateTestDB()
	if result := db.Create(&selectedItem); result.Error != nil {
		panic(result.Error)
	}

	itemRepo := NewItemRepository(db)

	err := itemRepo.DeleteSelectedItem(userId, channelId)
	assertNoError(err, t)

	_, err = itemRepo.GetSelectedItem(userId, channelId)
	assertError(err, t)
}

func TestGetItemByName(t *testing.T) {
	channelId := models.TwitchId("channel id")
	itemId := uuid.New()
	itemName := "item name"

	item := models.Item{
		ItemId: itemId,
		Name:   itemName,
	}

	channelItem := models.ChannelItem{
		ChannelId: channelId,
		ItemId:    itemId,
	}

	db := test.CreateTestDB()
	if result := db.Create(&item); result.Error != nil {
		panic(result.Error)
	}
	if result := db.Create(&channelItem); result.Error != nil {
		panic(result.Error)
	}

	itemRepo := NewItemRepository(db)
	got, err := itemRepo.GetItemByName(channelId, itemName)

	assertNoError(err, t)
	assertItemsEqual(got, item, t)
}

func TestGetItemById(t *testing.T) {
	itemId := uuid.New()
	item := models.Item{ItemId: itemId}

	db := test.CreateTestDB()
	if result := db.Create(&item); result.Error != nil {
		panic(result.Error)
	}

	itemRepo := NewItemRepository(db)
	got, err := itemRepo.GetItemById(itemId)

	assertNoError(err, t)
	assertItemsEqual(got, item, t)
}

func TestGetChannelsItems(t *testing.T) {
	channelId := models.TwitchId("channel id")
	itemId := uuid.New()

	item := models.Item{
		ItemId:  itemId,
		Name:    "item name",
		Rarity:  "rarity",
		Image:   "image",
		PrevImg: "prev image",
	}

	channelItem := models.ChannelItem{
		ChannelId: channelId,
		ItemId:    itemId,
	}

	db := test.CreateTestDB()
	if result := db.Create(&item); result.Error != nil {
		panic(result.Error)
	}
	if result := db.Create(&channelItem); result.Error != nil {
		panic(result.Error)
	}

	itemRepo := NewItemRepository(db)

	items, err := itemRepo.GetChannelsItems(channelId)
	assertNoError(err, t)

	expected := []models.Item{item}
	if !slices.Equal(items, expected) {
		t.Errorf("expected %s got %s", expected, items)
	}
}

func TestGetOwnedItems(t *testing.T) {
	channelId := models.TwitchId("channel id")
	userId := models.TwitchId("user id")

	itemId := uuid.New()
	item := models.Item{ItemId: itemId}

	owneditem := models.OwnedItem{
		UserId:    userId,
		ChannelId: channelId,
		ItemId:    itemId,
	}

	db := test.CreateTestDB()
	if result := db.Create(&item); result.Error != nil {
		panic(result.Error)
	}
	if result := db.Create(&owneditem); result.Error != nil {
		panic(result.Error)
	}

	itemRepo := NewItemRepository(db)

	items, err := itemRepo.GetOwnedItems(channelId, userId)
	assertNoError(err, t)

	expected := []models.Item{item}
	if !slices.Equal(items, expected) {
		t.Errorf("expected %s got %s", expected, items)
	}
}

func TestAddOwnedItem(t *testing.T) {
	channelId := models.TwitchId("channel id")
	userId := models.TwitchId("user id")
	itemId := uuid.New()
	transactionId := uuid.New()

	channelItem := models.ChannelItem{
		ItemId:    itemId,
		ChannelId: channelId,
	}

	db := test.CreateTestDB()
	if result := db.Create(&channelItem); result.Error != nil {
		panic(result.Error)
	}

	itemRepo := NewItemRepository(db)

	err := itemRepo.AddOwnedItem(userId, itemId, transactionId)

	assertNoError(err, t)
}

func TestCheckOwnedItem(t *testing.T) {
	t.Run("true when user owns item", func(t *testing.T) {
		userId := models.TwitchId("user id")
		itemId := uuid.New()

		ownedItem := models.OwnedItem{UserId: userId, ItemId: itemId}

		db := test.CreateTestDB()
		if result := db.Create(&ownedItem); result.Error != nil {
			panic(result.Error)
		}

		itemRepo := NewItemRepository(db)

		owned, err := itemRepo.CheckOwnedItem(userId, itemId)
		assertNoError(err, t)
		assertTrue(owned, t)
	})

	t.Run("false when item is unowned", func(t *testing.T) {
		userId := models.TwitchId("user id")
		itemId := uuid.New()

		db := test.CreateTestDB()

		itemRepo := NewItemRepository(db)

		owned, err := itemRepo.CheckOwnedItem(userId, itemId)
		assertNoError(err, t)
		assertTrue(!owned, t)
	})
}

func assertItemsEqual(got, want models.Item, t *testing.T) {
	t.Helper()

	if got != want {
		t.Errorf("expected %s got %s", want, got)
	}
}

func assertNoError(err error, t *testing.T) {
	t.Helper()

	if err != nil {
		t.Errorf("did not expect an error but received %s", err.Error())
	}
}

func assertError(err error, t *testing.T) {
	t.Helper()

	if err == nil {
		t.Error("expected an error but did not received one")
	}
}

func assertTrue(b bool, t *testing.T) {
	t.Helper()

	if !b {
		t.Errorf("expected true but received %t", b)
	}
}
