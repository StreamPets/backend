package repositories

import (
	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"gorm.io/gorm"
)

type ChannelRepo struct {
	db *gorm.DB
}

func NewChannelRepo(db *gorm.DB) *ChannelRepo {
	return &ChannelRepo{db: db}
}

func (repo *ChannelRepo) GetOverlayId(channelId models.UserId) (uuid.UUID, error) {
	var channel models.Channel

	if result := repo.db.Where("channel_id = ?", channelId).First(&channel); result.Error != nil {
		return uuid.UUID{}, result.Error
	}

	return channel.OverlayId, nil
}
