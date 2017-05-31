package simple

import (
	"github.com/VividCortex/ewma"
	"github.com/geoah/go-trade/market"
	"github.com/geoah/go-trade/strategy"
)

type simple struct {
	lastEMA float64
	ema     ewma.MovingAverage
}

// New random strategy
func New(window float64) (strategy.Strategy, error) {
	return &simple{
		ema: ewma.NewMovingAverage(window),
	}, nil
}

// Handle new candle
func (s *simple) Handle(candle *market.Candle) (strategy.Action, error) {
	// TODO Lock simple

	// find diff
	s.ema.Add(candle.Close)
	newEMA := s.ema.Value()
	diff := s.lastEMA - newEMA
	s.lastEMA = newEMA

	// report back
	if diff > 0 {
		return strategy.Buy, nil
	} else if diff < 0 {
		return strategy.Sell, nil
	}

	return strategy.Wait, nil
}
