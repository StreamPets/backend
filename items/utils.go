package items

import (
	"github.com/google/uuid"
	"github.com/streampets/backend/twitch"
)

type ErrSelectUnownedItem struct {
	UserId    twitch.Id
	ChannelId twitch.Id
	ItemId    uuid.UUID
}

func NewErrSelectUnownedItem(
	UserId twitch.Id,
	ChannelId twitch.Id,
	ItemId uuid.UUID,
) ErrSelectUnownedItem {
	return ErrSelectUnownedItem{
		UserId:    UserId,
		ChannelId: ChannelId,
		ItemId:    ItemId,
	}
}

func (e ErrSelectUnownedItem) Error() string {
	return "user tried to select an item they do not own"
}
