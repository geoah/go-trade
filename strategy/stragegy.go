package strategy

import (
	market "github.com/geoah/go-trade/market"
)

// Action -
type Action string

const (
	// Wait -
	Wait Action = "WAIT"
	// Buy -
	Buy Action = "BUY"
	// Sell -
	Sell Action = "SELL"
)

// Strategy -
type Strategy interface {
	Handle(candle *market.Candle) (Action, error)
}
