package market

import (
	"time"
)

// Aggregator -
type Aggregator interface {
	Aggregate(candle *Candle) error
	Notify(handler TradeHandler)
	Run()
}

// New -
func New(period time.Duration) (Aggregator, error) {
	agg := &aggregator{
		period:   period,
		handlers: []TradeHandler{},
	}
	return agg, nil
}

type aggregator struct {
	period   time.Duration
	handlers []TradeHandler
}

// Aggregate -
func (a *aggregator) Aggregate(candle *Candle) error {
	return nil
}

// Notify -
func (a *aggregator) Notify(handler TradeHandler) {
	a.handlers = append(a.handlers, handler)
}

// Run -
func (a *aggregator) Run() {
	// a.Listen()
}
