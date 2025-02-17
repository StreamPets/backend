package gorm

import "errors"

var ErrNoOverlayId = errors.New("no overlay id associated with the given channel id")

var ErrItemNotFound = errors.New("an item with the associated item id could not be found")
var ErrItemNotFoundByName = errors.New("an item with the associated item name could not be found")

var ErrAddItemNotExist = errors.New("tried to create an owned item for an item that does not exist")
