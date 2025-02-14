package database

import (
	"errors"

	"github.com/google/uuid"
	"github.com/streampets/backend/twitch"
)

var ErrNoOverlayId = errors.New("no overlay id associated with the given channel id")

type ErrItemNotFoundByName struct {
	ItemName string
}

func NewErrItemNotFoundByName(
	itemName string,
) ErrItemNotFoundByName {
	return ErrItemNotFoundByName{
		ItemName: itemName,
	}
}

func (e ErrItemNotFoundByName) Error() string {
	return "an item with the associated item name could not be found"
}

type ErrItemNotFoundById struct {
	ItemId uuid.UUID
}

func NewErrItemNotFoundById(
	itemId uuid.UUID,
) ErrItemNotFoundById {
	return ErrItemNotFoundById{
		ItemId: itemId,
	}
}

func (e ErrItemNotFoundById) Error() string {
	return "an item with the associated item id could not be found"
}

type ErrAddItemNotExist struct {
	UserId        twitch.Id
	ItemId        uuid.UUID
	TransactionId uuid.UUID
}

func NewErrAddItemNotExist(
	userId twitch.Id,
	itemId uuid.UUID,
	transactionId uuid.UUID,
) ErrAddItemNotExist {
	return ErrAddItemNotExist{
		UserId:        userId,
		ItemId:        itemId,
		TransactionId: transactionId,
	}
}

func (e ErrAddItemNotExist) Error() string {
	return "tried to add an owned item for an item that does not exist"
}
