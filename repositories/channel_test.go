package repositories

import (
	"testing"

	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/test"
	"github.com/streampets/backend/twitch"
	"github.com/stretchr/testify/assert"
)

func TestGetOverlayId(t *testing.T) {
	channelId := twitch.Id("channel id")
	overlayId := uuid.New()

	channel := models.Channel{
		ChannelId: channelId,
		OverlayId: overlayId,
	}

	db := test.CreateTestDB()
	if result := db.Create(&channel); result.Error != nil {
		panic(result.Error)
	}

	repo := NewChannelRepo(db)

	got, err := repo.GetOverlayId(channelId)

	assert.NoError(t, err)
	assert.Equal(t, overlayId, got)
}
