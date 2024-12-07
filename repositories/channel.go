package repositories

import (
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"gorm.io/gorm"
)

type ChannelRepository struct {
	db *gorm.DB
}

func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

func (repo *ChannelRepository) GetOverlayID(channelID models.TwitchID) (uuid.UUID, error) {
	var channel models.Channel

	if result := repo.db.Where("channel_id = ?", channelID).First(&channel); result.Error != nil {
		return uuid.UUID{}, result.Error
	}

	return channel.OverlayID, nil
}
