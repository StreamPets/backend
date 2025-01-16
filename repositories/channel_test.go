package repositories

import (
	"testing"

	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/test"
)

func TestGetOverlayId(t *testing.T) {
	channelId := models.UserId("channel id")
	overlayId := uuid.New()

	channel := models.Channel{
		ChannelId:   channelId,
		ChannelName: "channel name",
		OverlayId:   overlayId,
	}

	db := test.CreateTestDB()
	if result := db.Create(&channel); result.Error != nil {
		panic(result.Error)
	}

	repo := NewChannelRepo(db)

	got, err := repo.GetOverlayId(channelId)
	if err != nil {
		t.Errorf("did not expect an error but received: %s", err.Error())
	}

	if got != overlayId {
		t.Errorf("expected %s got %s", overlayId, got)
	}
}
