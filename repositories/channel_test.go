package repositories

import (
	"testing"

	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/test"
)

func TestGetOverlayID(t *testing.T) {
	channelID := models.TwitchID("channel id")
	overlayID := uuid.New()

	channel := models.Channel{
		ChannelID:   channelID,
		ChannelName: "channel name",
		OverlayID:   overlayID,
	}

	db := test.CreateTestDB()
	if result := db.Create(&channel); result.Error != nil {
		panic(result.Error)
	}

	repo := NewChannelRepo(db)

	got, err := repo.GetOverlayID(channelID)
	if err != nil {
		t.Errorf("did not expect an error but received: %s", err.Error())
	}

	if got != overlayID {
		t.Errorf("expected %s got %s", overlayID, got)
	}
}
