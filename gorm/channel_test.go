package gorm

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetOverlayId(t *testing.T) {
	channelId := "channel id"
	overlayId := uuid.New()

	channel := channel{
		ChannelId: channelId,
		OverlayId: overlayId,
	}

	db, err := CreateTestDB()
	assert.NoError(t, err)

	result := db.Create(&channel)
	assert.NoError(t, result.Error)

	database := NewChannelRepository(db)
	got, err := database.GetOverlayId(channelId)

	assert.NoError(t, err)
	assert.Equal(t, overlayId, got)
}
