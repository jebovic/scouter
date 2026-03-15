package watchlist

import "errors"

var (
	ErrNotFound      = errors.New("watchlist entry not found")
	ErrInvalidItemID = errors.New("invalid item ID")
)
