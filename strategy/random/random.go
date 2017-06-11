package random

import (
	"errors"
	"math/rand"
	"time"

	market "github.com/geoah/go-trade/market"
	strategy "github.com/geoah/go-trade/strategy"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type random struct {
	waitChance float64
	buyChance  float64
	sellChance float64
}

// New random strategy
func New(waitChance, buyChance, sellChance float64) (strategy.Strategy, error) {
	if (waitChance + buyChance + sellChance) != 1.0 {
		return nil, errors.New("Chances don't sum to 1.0")
	}
	return &random{
		waitChance: waitChance,
		buyChance:  buyChance,
		sellChance: sellChance,
	}, nil
}

// HandleCandle new candle
func (s *random) HandleCandle(candle *market.Candle) (market.Action, error) {
	r := rand.Float64()
	if r <= s.buyChance {
		return market.Buy, nil
	} else if r <= s.buyChance+s.sellChance {
		return market.Sell, nil
	} else if r <= s.buyChance+s.sellChance+s.waitChance {
		return market.Hold, nil
	}
	return market.Hold, errors.New("Woa, this should have NOT happened")
}
