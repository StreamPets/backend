package database

import (
	"errors"
)

var ErrNoOverlayId = errors.New("no overlay id associated with the given channel id")
var ErrItemNotFoundById = errors.New("an item with the associated item id could not be found")
var ErrAddItemNotExist = errors.New("tried to create an owned item for an item that does not exist")
var ErrItemNotFoundByName = errors.New("an item with the associated item name could not be found")
