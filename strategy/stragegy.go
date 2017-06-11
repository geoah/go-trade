package strategy

import (
	"errors"

	market "github.com/geoah/go-trade/market"
)

var (
	// ErrorNotImplemented -
	ErrorNotImplemented = errors.New("Not implemeneted")
)

// Strategy -
type Strategy interface {
	HandleCandle(candle *market.Candle) (market.Action, error)
}
