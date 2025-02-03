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

func (r *ChannelRepo) CreateChannel(channelId models.TwitchId, channelName string) (uuid.UUID, error) {
	channel := models.Channel{
		ChannelId:   channelId,
		ChannelName: channelName,
		OverlayId:   uuid.New(),
	}

	result := r.db.Create(&channel)
	if result.Error != nil {
		return uuid.UUID{}, result.Error
	}

	return channel.OverlayId, nil
}

func (r *ChannelRepo) CreateChannelItems(channelId models.TwitchId) ([]uuid.UUID, error) {
	greenItem := models.Item{
		ItemId:  uuid.New(),
		Name:    "green",
		Rarity:  models.Common,
		Image:   "./assets/green-rex.png",
		PrevImg: "./assets/green-rex-prev.png",
	}

	redItem := models.Item{
		ItemId:  uuid.New(),
		Name:    "red",
		Rarity:  models.Common,
		Image:   "./assets/red-rex.png",
		PrevImg: "./assets/red-rex-prev.png",
	}

	blueItem := models.Item{
		ItemId:  uuid.New(),
		Name:    "blue",
		Rarity:  models.Common,
		Image:   "./assets/blue-rex.png",
		PrevImg: "./assets/blue-rex-prev.png",
	}

	if err := r.db.Transaction(func(tx *gorm.DB) error {
		if result := r.db.Create(&greenItem); result.Error != nil {
			return result.Error
		}
		if result := r.db.Create(&redItem); result.Error != nil {
			return result.Error
		}
		if result := r.db.Create(&blueItem); result.Error != nil {
			return result.Error
		}
		return nil
	}); err != nil {
		return []uuid.UUID{}, err
	}

	return []uuid.UUID{
		greenItem.ItemId,
		redItem.ItemId,
		blueItem.ItemId,
	}, nil
}

func (r *ChannelRepo) CreateDefaultItem(channelId models.TwitchId, itemId uuid.UUID) error {
	return r.db.Create(&models.DefaultChannelItem{
		ChannelId: channelId,
		ItemId:    itemId,
	}).Error
}
