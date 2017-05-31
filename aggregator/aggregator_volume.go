package aggregator

import (
	"sort"

	"github.com/geoah/go-trade/market"
	"github.com/thetruetrade/gotrade"
)

// Volume -
type Volume struct {
	volumeLimit   float64
	volumeCurrent float64

	trades []*market.Trade

	streamBarIndex int

	handlers       []market.CandleHandler
	handlersDOHLCV []gotrade.DOHLCVTickReceiver
}

// NewVolumeAggregator -
func NewVolumeAggregator(volume float64) (Aggregator, error) {
	agg := &Volume{
		volumeLimit: volume,
		handlers:    []market.CandleHandler{},
	}
	return agg, nil
}

// Handle -
func (a *Volume) Handle(trade *market.Trade) error {
	a.volumeCurrent += trade.Size
	a.trades = append(a.trades, trade)
	if a.volumeCurrent >= a.volumeLimit {
		a.tick(trade)
	}
	return nil
}

func (a *Volume) tick(trade *market.Trade) {
	// sort trades
	sort.Slice(a.trades, func(i, j int) bool {
		return a.trades[i].Time.Before(a.trades[j].Time)
	})
	// create a new candle with random h/l
	c := &market.Candle{
		Time:  a.trades[0].Time,
		Open:  a.trades[0].Price,
		Close: a.trades[len(a.trades)-1].Price,
		High:  a.trades[0].Price,
		Low:   a.trades[0].Price,
	}
	// go through trades and find h/l
	for _, trade := range a.trades {
		if trade.Price > c.High {
			c.High = trade.Price
		}
		if trade.Price < c.Low {
			c.Low = trade.Price
		}
		c.Volume += trade.Size
	}

	// notify
	a.notify(c)

	// clear trades
	a.trades = []*market.Trade{}
}

// Register -
func (a *Volume) Register(handler market.CandleHandler) {
	a.handlers = append(a.handlers, handler)
}

// AddTickSubscription -
func (a *Volume) AddTickSubscription(handler gotrade.DOHLCVTickReceiver) {
	a.handlersDOHLCV = append(a.handlersDOHLCV, handler)
}

func (a *Volume) notify(candle *market.Candle) {
	for _, h := range a.handlers {
		h.Handle(candle) // TODO Handle error
	}
	dohlcv := TradeToDOHLCV(candle)
	for _, h := range a.handlersDOHLCV {
		h.ReceiveDOHLCVTick(dohlcv, a.streamBarIndex)
	}
	a.streamBarIndex++ // TODO Not sure about what streamBarIndex does
}

// Run -
func (a *Volume) Run() {
	// a.Listen()
}
