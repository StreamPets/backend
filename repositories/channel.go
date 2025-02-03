package repositories

import (
	"errors"

	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"gorm.io/gorm"
)

var ErrNoOverlayId = errors.New("no overlay id found")

type ChannelRepo struct {
	db *gorm.DB
}

func NewChannelRepo(db *gorm.DB) *ChannelRepo {
	return &ChannelRepo{db: db}
}

func (r *ChannelRepo) GetOverlayId(channelId models.TwitchId) (uuid.UUID, error) {
	var channel models.Channel

	if result := r.db.Where("channel_id = ?", channelId).First(&channel); result.Error == gorm.ErrRecordNotFound {
		return uuid.UUID{}, ErrNoOverlayId
	} else if result.Error != nil {
		return uuid.UUID{}, result.Error
	}

	return channel.OverlayId, nil
}
