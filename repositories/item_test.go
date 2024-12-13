package repositories_test

import (
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/repositories"
	"github.com/streampets/backend/test"
	"gorm.io/gorm"
)

func TestGetSelectedItem(t *testing.T) {
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

	itemRepo := repositories.NewItemRepository(db)
	got, err := itemRepo.GetSelectedItem(userID, channelID)

	assertNoError(err, t)
	assertItemsEqual(got, item, t)
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

	itemRepo := repositories.NewItemRepository(db)
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

	itemRepo := repositories.NewItemRepository(db)
	got, err := itemRepo.GetItemByID(itemID)

	assertNoError(err, t)
	assertItemsEqual(got, item, t)
}

func TestGetScheduledItems(t *testing.T) {
	channelID := models.TwitchID("channel id")
	dayOfWeek := models.Monday
	itemID := uuid.New()

	item := models.Item{
		ItemID:  itemID,
		Name:    "item name",
		Rarity:  "rarity",
		Image:   "image",
		PrevImg: "prev image",
	}

	schedule := models.Schedule{
		ScheduleID: uuid.New(),
		DayOfWeek:  dayOfWeek,
		ItemID:     itemID,
		ChannelID:  channelID,
	}

	db := test.CreateTestDB()
	if result := db.Create(&item); result.Error != nil {
		panic(result.Error)
	}
	if result := db.Create(&schedule); result.Error != nil {
		panic(result.Error)
	}

	itemRepo := repositories.NewItemRepository(db)

	items, err := itemRepo.GetScheduledItems(channelID, dayOfWeek)
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

	itemRepo := repositories.NewItemRepository(db)

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

	itemRepo := repositories.NewItemRepository(db)

	err := itemRepo.AddOwnedItem(userID, itemID, transactionID)

	assertNoError(err, t)
}

func TestCheckOwnedItem(t *testing.T) {
	t.Run("no error when user owns item", func(t *testing.T) {
		userID := models.TwitchID("user id")
		itemID := uuid.New()

		ownedItem := models.OwnedItem{UserID: userID, ItemID: itemID}

		db := test.CreateTestDB()
		if result := db.Create(&ownedItem); result.Error != nil {
			panic(result.Error)
		}

		itemRepo := repositories.NewItemRepository(db)

		err := itemRepo.CheckOwnedItem(userID, itemID)
		assertNoError(err, t)
	})

	t.Run("error when item is unowned", func(t *testing.T) {
		userID := models.TwitchID("user id")
		itemID := uuid.New()

		db := test.CreateTestDB()

		itemRepo := repositories.NewItemRepository(db)

		err := itemRepo.CheckOwnedItem(userID, itemID)
		if err == nil {
			t.Errorf("expected an error but did not receive one")
		} else if err != gorm.ErrRecordNotFound {
			t.Errorf("expected %s got %s", gorm.ErrRecordNotFound, err)
		}
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
