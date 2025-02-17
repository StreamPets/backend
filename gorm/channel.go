package gorm

import (
	"github.com/google/uuid"
)

type ChannelRepository struct {
	db *DB
}

func NewChannelRepository(db *DB) *ChannelRepository {
	return &ChannelRepository{
		db: db,
	}
}

func (r *ChannelRepository) GetOverlayId(channelId string) (uuid.UUID, error) {
	var channel channel

	result := r.db.Where("channel_id = ?", channelId).First(&channel)
	if result.Error != nil {
		return uuid.UUID{}, result.Error
	}

	return channel.OverlayId, nil
}
