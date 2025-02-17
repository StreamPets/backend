package items

import (
	"errors"
)

var ErrSelectUnownedItem = errors.New("user tried to select an item they do not own")
