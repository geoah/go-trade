package aggregator

import (
	"sort"
	"time"

	"github.com/thetruetrade/gotrade"

	market "github.com/geoah/go-trade/market"
)

// Time -
type Time struct {
	nextTickStart time.Time
	period        time.Duration
	trades        []*market.Trade
	empty         bool

	streamBarIndex int

	handlers       []market.CandleHandler
	handlersDOHLCV []gotrade.DOHLCVTickReceiver
}

// NewTimeAggregator -
func NewTimeAggregator(period time.Duration) (Aggregator, error) {
	agg := &Time{
		period:        period,
		handlers:      []market.CandleHandler{},
		nextTickStart: time.Time{}.UTC(),
		empty:         true,
	}
	return agg, nil
}

func (a *Time) isInCurrentTick(trade *market.Trade) bool {
	start := a.nextTickStart.Add(-a.period)
	end := a.nextTickStart
	if trade.Time.After(start) && trade.Time.Before(end) {
		return true
	}
	return false
}

func (a *Time) isInFutureTick(trade *market.Trade) bool {
	start := a.nextTickStart
	// end := a.nextTickStart.Add(a.period)
	if trade.Time.After(start) { //} && trade.Time.Before(end) {
		return true
	}
	return false
}

// Handle -
func (a *Time) HandleTrade(trade *market.Trade) error {
	if a.empty {
		a.empty = false
		a.tick(trade)
		return nil
	}
	if a.isInCurrentTick(trade) {
		a.trades = append(a.trades, trade)
	} else if a.isInFutureTick(trade) {
		a.tick(trade)
	}
	return nil
}

func (a *Time) tick(trade *market.Trade) {
	if len(a.trades) > 0 {
		// sort trades
		sort.Slice(a.trades, func(i, j int) bool {
			return a.trades[i].Time.Before(a.trades[j].Time)
		})
		// create a new candle with random h/l
		c := &market.Candle{
			Time:  a.nextTickStart.Add(-a.period),
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
	}
	// mark start of next tick
	a.nextTickStart = trade.Time.UTC().Round(a.period).Add(a.period)
	// clear trades
	a.trades = []*market.Trade{trade}
}

// Register -
func (a *Time) Register(handler market.CandleHandler) {
	a.handlers = append(a.handlers, handler)
}

// AddTickSubscription -
func (a *Time) AddTickSubscription(handler gotrade.DOHLCVTickReceiver) {
	a.handlersDOHLCV = append(a.handlersDOHLCV, handler)
}

func (a *Time) notify(candle *market.Candle) {
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
func (a *Time) Run() {
	// a.Listen()
}

func TradeToDOHLCV(c *market.Candle) gotrade.DOHLCV {
	return gotrade.NewDOHLCVDataItem(c.Time, c.Open, c.High, c.Low, c.Close, c.Volume)
}
