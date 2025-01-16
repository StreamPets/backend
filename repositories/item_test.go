package repositories

import (
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/test"
)

func TestGetSelectedItem(t *testing.T) {
	t.Run("selected item returned when item selected", func(t *testing.T) {
		channelID := models.TwitchID("channel id")
		userID := models.TwitchID("user id")

		itemID := uuid.New()
		item := models.Item{ItemID: itemID}

		selectedItem := models.SelectedItem{
			UserID:    userID,
			ChannelID: channelID,
			ItemID:    itemID,
		}

		db := test.CreateTestDB()
		if result := db.Create(&item); result.Error != nil {
			panic(result.Error)
		}
		if result := db.Create(&selectedItem); result.Error != nil {
			panic(result.Error)
		}

		itemRepo := NewItemRepository(db)
		got, err := itemRepo.GetSelectedItem(userID, channelID)

		assertNoError(err, t)
		assertItemsEqual(got, item, t)
	})

	t.Run("default item returned when no item selected", func(t *testing.T) {
		channelID := models.TwitchID("channel id")
		userID := models.TwitchID("user id")

		itemID := uuid.New()
		item := models.Item{ItemID: itemID}

		defaultItem := models.DefaultChannelItem{
			ItemID:    itemID,
			ChannelID: channelID,
		}

		db := test.CreateTestDB()
		if result := db.Create(&item); result.Error != nil {
			panic(result.Error)
		}
		if result := db.Create(&defaultItem); result.Error != nil {
			panic(result.Error)
		}

		itemRepo := NewItemRepository(db)
		got, err := itemRepo.GetSelectedItem(userID, channelID)

		assertNoError(err, t)
		assertItemsEqual(got, item, t)
	})
}

func TestSetSelectedItem(t *testing.T) {
	channelID := models.TwitchID("channel id")
	userID := models.TwitchID("user id")

	itemID := uuid.New()
	item := models.Item{ItemID: itemID}

	newItemID := uuid.New()
	newItem := models.Item{ItemID: newItemID}

	selectedItem := models.SelectedItem{
		UserID:    userID,
		ChannelID: channelID,
		ItemID:    itemID,
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

	got, err := itemRepo.GetSelectedItem(userID, channelID)
	assertNoError(err, t)
	assertItemsEqual(got, item, t)

	err = itemRepo.SetSelectedItem(userID, channelID, newItemID)
	assertNoError(err, t)

	got, err = itemRepo.GetSelectedItem(userID, channelID)
	assertNoError(err, t)
	assertItemsEqual(got, newItem, t)
}

func TestGetItemByName(t *testing.T) {
	channelID := models.TwitchID("channel id")
	itemID := uuid.New()
	itemName := "item name"

	item := models.Item{
		ItemID: itemID,
		Name:   itemName,
	}

	channelItem := models.ChannelItem{
		ChannelID: channelID,
		ItemID:    itemID,
	}

	db := test.CreateTestDB()
	if result := db.Create(&item); result.Error != nil {
		panic(result.Error)
	}
	if result := db.Create(&channelItem); result.Error != nil {
		panic(result.Error)
	}

	itemRepo := NewItemRepository(db)
	got, err := itemRepo.GetItemByName(channelID, itemName)

	assertNoError(err, t)
	assertItemsEqual(got, item, t)
}

func TestGetItemByID(t *testing.T) {
	itemID := uuid.New()
	item := models.Item{ItemID: itemID}

	db := test.CreateTestDB()
	if result := db.Create(&item); result.Error != nil {
		panic(result.Error)
	}

	itemRepo := NewItemRepository(db)
	got, err := itemRepo.GetItemByID(itemID)

	assertNoError(err, t)
	assertItemsEqual(got, item, t)
}

func TestGetChannelsItems(t *testing.T) {
	channelID := models.TwitchID("channel id")
	itemID := uuid.New()

	item := models.Item{
		ItemID:  itemID,
		Name:    "item name",
		Rarity:  "rarity",
		Image:   "image",
		PrevImg: "prev image",
	}

	channelItem := models.ChannelItem{
		ChannelID: channelID,
		ItemID:    itemID,
	}

	db := test.CreateTestDB()
	if result := db.Create(&item); result.Error != nil {
		panic(result.Error)
	}
	if result := db.Create(&channelItem); result.Error != nil {
		panic(result.Error)
	}

	itemRepo := NewItemRepository(db)

	items, err := itemRepo.GetChannelsItems(channelID)
	assertNoError(err, t)

	expected := []models.Item{item}
	if !slices.Equal(items, expected) {
		t.Errorf("expected %s got %s", expected, items)
	}
}

func TestGetOwnedItems(t *testing.T) {
	channelID := models.TwitchID("channel id")
	userID := models.TwitchID("user id")
	itemID := uuid.New()

	item := models.Item{
		ItemID:  itemID,
		Name:    "item name",
		Rarity:  "rarity",
		Image:   "image",
		PrevImg: "prev image",
	}

	owneditem := models.OwnedItem{
		UserID:    "user id",
		ChannelID: "channel id",
		ItemID:    itemID,
	}

	db := test.CreateTestDB()
	if result := db.Create(&item); result.Error != nil {
		panic(result.Error)
	}
	if result := db.Create(&owneditem); result.Error != nil {
		panic(result.Error)
	}

	itemRepo := NewItemRepository(db)

	items, err := itemRepo.GetOwnedItems(channelID, userID)
	assertNoError(err, t)

	expected := []models.Item{item}
	if !slices.Equal(items, expected) {
		t.Errorf("expected %s got %s", expected, items)
	}
}

func TestAddOwnedItem(t *testing.T) {
	channelID := models.TwitchID("channel id")
	userID := models.TwitchID("user id")
	itemID := uuid.New()
	transactionID := uuid.New()

	channelItem := models.ChannelItem{
		ItemID:    itemID,
		ChannelID: channelID,
	}

	db := test.CreateTestDB()
	if result := db.Create(&channelItem); result.Error != nil {
		panic(result.Error)
	}

	itemRepo := NewItemRepository(db)

	err := itemRepo.AddOwnedItem(userID, itemID, transactionID)

	assertNoError(err, t)
}

func TestCheckOwnedItem(t *testing.T) {
	t.Run("true when user owns item", func(t *testing.T) {
		userID := models.TwitchID("user id")
		itemID := uuid.New()

		ownedItem := models.OwnedItem{UserID: userID, ItemID: itemID}

		db := test.CreateTestDB()
		if result := db.Create(&ownedItem); result.Error != nil {
			panic(result.Error)
		}

		itemRepo := NewItemRepository(db)

		owned, err := itemRepo.CheckOwnedItem(userID, itemID)
		assertNoError(err, t)
		assertTrue(owned, t)
	})

	t.Run("false when item is unowned", func(t *testing.T) {
		userID := models.TwitchID("user id")
		itemID := uuid.New()

		db := test.CreateTestDB()

		itemRepo := NewItemRepository(db)

		owned, err := itemRepo.CheckOwnedItem(userID, itemID)
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

func assertTrue(b bool, t *testing.T) {
	t.Helper()

	if !b {
		t.Errorf("expected true but received %t", b)
	}
}
