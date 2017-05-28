package fake

import (
	"errors"
	"sort"
	"sync"

	market "github.com/geoah/go-trade/market"
)

type Fake struct {
	sync.Mutex
	handlers []market.Handler
	candles  []*market.Candle
	asset    float64
	currency float64
}

func New(candles []*market.Candle, asset, currency float64) (market.Market, error) {
	m := &Fake{
		handlers: []market.Handler{},
		candles:  candles,
		asset:    asset,
		currency: currency,
	}
	sort.Slice(m.candles, func(i, j int) bool {
		return m.candles[i].Time.Before(m.candles[j].Time)
	})
	return m, nil
}

func (m *Fake) Notify(handler market.Handler) {
	m.handlers = append(m.handlers, handler)
}

func (m *Fake) Buy(quantity, price float64) error {
	m.Lock()
	defer m.Unlock()
	cost := quantity * price
	if cost > m.currency {
		return errors.New("Not enough currency")
	}
	m.currency -= cost
	m.asset += quantity
	return nil
}

func (m *Fake) Sell(quantity, price float64) error {
	m.Lock()
	defer m.Unlock()
	if quantity > m.asset {
		return errors.New("Not enough assets")
	}
	m.asset -= quantity
	m.currency += quantity * price
	return nil
}

func (m *Fake) GetBalance() (assets float64, currency float64, err error) {
	m.Lock()
	defer m.Unlock()
	return m.asset, m.currency, nil
}

func (m *Fake) Run() {
	for i := range m.candles {
		for _, h := range m.handlers {
			if h != nil {
				h(m.candles[i])
			}
		}
	}
}
