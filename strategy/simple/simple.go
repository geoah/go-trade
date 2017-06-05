package simple

import (
	ewma "github.com/VividCortex/ewma"

	market "github.com/geoah/go-trade/market"
	strategy "github.com/geoah/go-trade/strategy"
)

type simple struct {
	ema ewma.MovingAverage

	lastEma     float64
	lastEmaDiff float64
}

// New random strategy
func New(window float64) (strategy.Strategy, error) {
	return &simple{
		ema: ewma.NewMovingAverage(window),
	}, nil
}

// Handle new candle
func (s *simple) Handle(candle *market.Candle) (strategy.Action, error) {
	// add candle to our ema
	s.ema.Add(candle.Close)

	// save the ema in the candle
	candle.Ema = s.ema.Value()

	// set wait as fallback
	act := strategy.Wait

	// get the direction of the trend
	diff := s.lastEma - candle.Ema

	// if the direction has changed since our last tick
	if diff > 0 && s.lastEmaDiff < 0 {
		// from down to up, then buy
		act = strategy.Buy
	} else if diff < 0 && s.lastEmaDiff > 0 {
		// from up to down, then sell
		act = strategy.Sell
	}

	// store new ema and direction
	s.lastEma = candle.Ema
	s.lastEmaDiff = diff

	// suggest action
	return act, nil
}
