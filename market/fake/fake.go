package fake

import (
	"errors"
	"fmt"
	"sync"
	"time"

	market "github.com/geoah/go-trade/market"
	persistence "github.com/geoah/go-trade/persistence"
)

const (
	// Name -
	Name = "fake"
)

// Fake -
type Fake struct {
	sync.Mutex
	persistence    persistence.Persistence
	handlers       []market.TradeHandler
	updateHandlers []market.UpdateHandler
	asset          float64
	currency       float64
	marketName     string
	productName    string
	feesPercent    float64
	start          time.Time
	end            time.Time
}

// New -
func New(pe persistence.Persistence, mrk, prd string, start, end time.Time, asset, currency float64) (market.Market, error) {
	m := &Fake{
		handlers:       []market.TradeHandler{},
		updateHandlers: []market.UpdateHandler{},
		asset:          asset,
		currency:       currency,
		persistence:    pe,
		start:          start,
		end:            end,
		marketName:     mrk,
		productName:    prd,
		feesPercent:    1.0,
	}
	return m, nil
}

// RegisterForTrades -
func (m *Fake) RegisterForTrades(handler market.TradeHandler) {
	m.handlers = append(m.handlers, handler)
}

// RegisterForUpdates -
func (m *Fake) RegisterForUpdates(handler market.UpdateHandler) {
	m.updateHandlers = append(m.updateHandlers, handler)
}

// Buy -
func (m *Fake) Buy(size, price float64) error {
	cost := size * price * m.feesPercent // TODO Check fees
	if cost > m.currency {
		return errors.New("Not enough currency")
	}

	m.Lock()
	m.currency -= cost
	m.asset += size
	m.Unlock()

	// push update
	upd := &market.Update{
		Action: market.Buy,
		Price:  price,
		Size:   size,
		Time:   time.Now(),
	}
	for _, h := range m.updateHandlers {
		if h != nil {
			h.HandleUpdate(upd) // TODO Handle error
		}
	}

	return nil
}

// Sell -
func (m *Fake) Sell(size, price float64) error {
	if size > m.asset {
		return errors.New("Not enough assets")
	}

	m.Lock()
	m.asset -= size
	m.currency += size * price / m.feesPercent // TODO Check fees
	m.Unlock()

	// push update
	upd := &market.Update{
		Action: market.Sell,
		Price:  price,
		Size:   size,
		Time:   time.Now(),
	}
	for _, h := range m.updateHandlers {
		if h != nil {
			h.HandleUpdate(upd) // TODO Handle error
		}
	}

	return nil
}

// GetBalance -
func (m *Fake) GetBalance() (assets float64, currency float64, err error) {
	m.Lock()
	defer m.Unlock()
	return m.asset, m.currency, nil
}

// Run -
func (m *Fake) Run() {
	// TODO make this async and send smaller batches
	trades, err := m.persistence.GetTrades(m.marketName, m.productName, m.start, m.end)
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

// Backfill -
func (m *Fake) Backfill(end time.Time) error {
	return errors.New("Not implemented")
}
