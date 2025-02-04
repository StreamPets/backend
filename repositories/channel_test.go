package repositories

import (
	"testing"

	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"github.com/streampets/backend/test"
	"github.com/stretchr/testify/assert"
)

func TestGetOverlayId(t *testing.T) {
	channelId := models.TwitchId("channel id")
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

func TestCreateChannel(t *testing.T) {
	channelId := models.TwitchId("channel id")
	channelName := "channel name"

	db := test.CreateTestDB()
	repo := NewChannelRepo(db)

	expected, err := repo.CreateChannel(channelId, channelName)
	assert.NoError(t, err)

	actual, err := repo.GetOverlayId(channelId)
	assert.NoError(t, err)

	assert.Equal(t, expected, actual)
}
