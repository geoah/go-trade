package aggregator

import (
	"github.com/thetruetrade/gotrade"

	market "github.com/geoah/go-trade/market"
)

// Aggregator implements DOHLCVStreamSubscriber
type Aggregator interface {
	// Run aggregator
	Run()
	// Register market.CandleHandler
	Register(handler market.CandleHandler)
	// Handle implements market.TradeHandler
	Handle(trade *market.Trade) error
	// AddTickSubscription is the same as our Notify but for gotrade
	AddTickSubscription(subscriber gotrade.DOHLCVTickReceiver)
}
