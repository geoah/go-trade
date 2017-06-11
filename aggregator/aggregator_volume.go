package aggregator

import (
	"sort"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/thetruetrade/gotrade"

	market "github.com/geoah/go-trade/market"
)

var (
	timeNil = time.Unix(0, 0)
)

// Volume -
type Volume struct {
	tradesMin     int
	tradesMax     int
	volumeLimit   float64
	volumeCurrent float64
	volumeReset   time.Time

	trades []*market.Trade

	streamBarIndex int

	handlers       []market.CandleHandler
	handlersDOHLCV []gotrade.DOHLCVTickReceiver
}

// NewVolumeAggregator -
func NewVolumeAggregator(volume float64) (Aggregator, error) {
	agg := &Volume{
		volumeLimit: volume,
		tradesMin:   1, // TODO Make configurable
		tradesMax:   5, // TODO Make configurable
		volumeReset: timeNil,
		handlers:    []market.CandleHandler{},
	}
	return agg, nil
}

// Handle -
func (a *Volume) HandleTrade(trade *market.Trade) error {
	if a.volumeReset == timeNil {
		a.volumeReset = trade.Time
	}
	a.volumeCurrent += trade.Size
	a.trades = append(a.trades, trade)
	if a.volumeCurrent >= a.volumeLimit && len(a.trades) > a.tradesMin {
		a.tick(trade)
	} else if len(a.trades) > a.tradesMax {
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

	// log stats
	diff := trade.Time.Sub(a.volumeReset)
	logrus.
		WithField("second", diff.Seconds()).
		WithField("volume", a.volumeLimit).
		WithField("trades", len(a.trades)).
		Debugf("Volume filled")

	// clear trades
	a.trades = []*market.Trade{}

	// reset time
	a.volumeReset = trade.Time

	// reset volume
	a.volumeCurrent = 0
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
		h.HandleCandle(candle) // TODO Handle error
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
