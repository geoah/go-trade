package persistence

import (
	"time"

	market "github.com/geoah/go-trade/market"
)

// Persistence -
type Persistence interface {
	PutTrade(trades ...*market.Trade) error
	GetTrades(mrk, prd string, start, end time.Time) ([]*market.Trade, error)
}
