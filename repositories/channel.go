package repositories

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/streampets/backend/models"
	"gorm.io/gorm"
)

type ErrNoOverlayId struct {
	ChannelId models.TwitchId
}

func (e *ErrNoOverlayId) Error() string {
	return fmt.Sprintf("no overlay id associated with the channel id %s", e.ChannelId)
}

func NewErrNoOverlayId(channelId models.TwitchId) error {
	return &ErrNoOverlayId{ChannelId: channelId}
}

type ChannelRepo struct {
	db *gorm.DB
}

func NewChannelRepo(db *gorm.DB) *ChannelRepo {
	return &ChannelRepo{db: db}
}

func (r *ChannelRepo) GetOverlayId(channelId models.TwitchId) (uuid.UUID, error) {
	var channel models.Channel

	if result := r.db.Where("channel_id = ?", channelId).First(&channel); result.Error == gorm.ErrRecordNotFound {
		return uuid.UUID{}, NewErrNoOverlayId(channelId)
	} else if result.Error != nil {
		return uuid.UUID{}, result.Error
	}

	return channel.OverlayId, nil
}
