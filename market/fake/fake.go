package fake

import (
	"errors"
	"fmt"
	"sync"
	"time"

	market "github.com/geoah/go-trade/market"
	persistence "github.com/geoah/go-trade/persistence"
	"github.com/geoah/go-trade/utils"
	"github.com/sirupsen/logrus"
)

const (
	Name = "fake"
)

type Fake struct {
	sync.Mutex
	persistence persistence.Persistence
	handlers    []market.TradeHandler
	asset       float64
	currency    float64
	back        time.Duration
	marketName  string
	productName string
	feesPercent float64
}

func New(pe persistence.Persistence, mrk, prd string, back time.Duration, asset, currency float64) (market.Market, error) {
	m := &Fake{
		handlers:    []market.TradeHandler{},
		asset:       asset,
		currency:    currency,
		persistence: pe,
		back:        back,
		marketName:  mrk,
		productName: prd,
		feesPercent: 1.0,
	}
	return m, nil
}

func (m *Fake) RegisterForTrades(handler market.TradeHandler) {
	m.handlers = append(m.handlers, handler)
}

func (m *Fake) RegisterForUpdates(handler market.UpdateHandler) {
}

func (m *Fake) Buy(quantity, price float64) error {
	m.Lock()
	defer m.Unlock()
	logrus.
		WithField("price", utils.TrimFloat64(price, 2)).
		WithField("size", utils.TrimFloat64(quantity, 8)).
		Infof("Placed buy order")
	cost := quantity * price * m.feesPercent // TODO Check fees
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
	logrus.
		WithField("price", utils.TrimFloat64(price, 2)).
		WithField("size", utils.TrimFloat64(quantity, 8)).
		Infof("Placed sell order")
	if quantity > m.asset {
		return errors.New("Not enough assets")
	}
	m.asset -= quantity
	m.currency += quantity * price / m.feesPercent // TODO Check fees
	return nil
}

func (m *Fake) GetBalance() (assets float64, currency float64, err error) {
	m.Lock()
	defer m.Unlock()
	return m.asset, m.currency, nil
}

func (m *Fake) Run() {
	end := time.Now()
	start := end.Add(-m.back)
	// TODO make this async and send smaller batches
	trades, err := m.persistence.GetTrades(m.marketName, m.productName, start, end)
	if err != nil {
		fmt.Println("Could not get trades", err)
		return
	}
	if len(trades) == 0 {
		fmt.Println("No trades for the given duration, you might want to backfill first.")
		fmt.Println("eg. go-trade backfill --days 5")
		return
	}
	for _, trade := range trades {
		for _, h := range m.handlers {
			if h != nil {
				h.HandleTrade(trade) // TODO Handle error
			}
		}
	}
}

func (m *Fake) Backfill(end time.Time) error {
	return errors.New("Not implemented")
}
