package strategy

import (
	"errors"

	market "github.com/geoah/go-trade/market"
)

var (
	ErrorNotImplemented = errors.New("Not implemeneted")
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
