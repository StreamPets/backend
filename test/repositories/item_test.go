package repositories_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/repositories"
)

func TestGetSelectedItem(t *testing.T) {
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

	selectedItem := models.SelectedItem{
		UserID:    userID,
		ChannelID: channelID,
		ItemID:    itemID,
	}

	db := createTestDB()
	if result := db.Create(&item); result.Error != nil {
		panic(result.Error)
	}
	if result := db.Create(&selectedItem); result.Error != nil {
		panic(result.Error)
	}

	itemRepo := repositories.NewItemRepository(db)
	got, err := itemRepo.GetSelectedItem(userID, channelID)

	assertNoError(err, t)
	assertItemsEqual(got, item, t)
}

func TestSetSelectedItem(t *testing.T) {
	channelID := models.TwitchID("channel id")
	userID := models.TwitchID("user id")
	newItemID := uuid.New()
	itemID := uuid.New()

	item := models.Item{
		ItemID:  itemID,
		Name:    "name",
		Rarity:  "rarity",
		Image:   "image",
		PrevImg: "prev image",
	}

	newItem := models.Item{
		ItemID:  newItemID,
		Name:    "new name",
		Rarity:  "new rarity",
		Image:   "new image",
		PrevImg: "new prev image",
	}

	selectedItem := models.SelectedItem{
		UserID:    userID,
		ChannelID: channelID,
		ItemID:    itemID,
	}

	db := createTestDB()
	if result := db.Create(&item); result.Error != nil {
		panic(result.Error)
	}
	if result := db.Create(&newItem); result.Error != nil {
		panic(result.Error)
	}
	if result := db.Create(&selectedItem); result.Error != nil {
		panic(result.Error)
	}

	itemRepo := repositories.NewItemRepository(db)

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
		ItemID:  itemID,
		Name:    itemName,
		Rarity:  "rarity",
		Image:   "image",
		PrevImg: "prev image",
	}

	channelItem := models.ChannelItem{
		ChannelID: channelID,
		ItemID:    itemID,
	}

	db := createTestDB()
	if result := db.Create(&item); result.Error != nil {
		panic(result.Error)
	}
	if result := db.Create(&channelItem); result.Error != nil {
		panic(result.Error)
	}

	itemRepo := repositories.NewItemRepository(db)
	got, err := itemRepo.GetItemByName(channelID, itemName)

	assertNoError(err, t)
	assertItemsEqual(got, item, t)
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
