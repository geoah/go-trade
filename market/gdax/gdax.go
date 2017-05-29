package gdax

import (
	"errors"
	"fmt"
	"os"
	"time"

	ws "github.com/gorilla/websocket"
	exchange "github.com/preichenberger/go-coinbase-exchange"

	market "github.com/geoah/go-trade/market"
	persistence "github.com/geoah/go-trade/persistence"
)

const (
	// Name of the market
	Name = "gdax"
)

// gdax -
type gdax struct {
	product     string
	handlers    []market.TradeHandler
	client      *exchange.Client
	persistence persistence.Persistence
}

// New gdax market
func New(persistence persistence.Persistence, product string) (market.Market, error) {
	secret := os.Getenv("COINBASE_SECRET")
	key := os.Getenv("COINBASE_KEY")
	passphrase := os.Getenv("COINBASE_PASSPHRASE")

	mrk := &gdax{
		product:     product,
		handlers:    []market.TradeHandler{},
		client:      exchange.NewClient(secret, key, passphrase),
		persistence: persistence,
	}

	return mrk, nil
}

func (m *gdax) Listen() {
	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.gdax.com", nil)
	if err != nil {
		println(err.Error())
	}
	subscribe := map[string]string{
		"type":       "subscribe",
		"product_id": m.product,
	}
	if err := wsConn.WriteJSON(subscribe); err != nil {
		println("gdax ws sub error", err.Error())
	}
	message := exchange.Message{}
	for true {
		if err := wsConn.ReadJSON(&message); err != nil {
			println("gdax ws read error", err.Error())
			break
		}
		if message.Type == "match" && message.Reason == "filled" {
			c := &market.Trade{
				ID:      fmt.Sprintf("%s.%s.%d", Name, m.product, message.TradeId),
				Market:  Name,
				Product: m.product,
				TradeID: message.TradeId,
				Price:   message.Price,
				Size:    message.Size,
				Time:    message.Time.Time(),
				Side:    message.Side,
			}
			// TODO move to channels
			for _, h := range m.handlers {
				if h != nil {
					h(c)
				}
			}
		}
	}
}

// Notify -
func (m *gdax) Notify(handler market.TradeHandler) {
	m.handlers = append(m.handlers, handler)
}

// Buy -
func (m *gdax) Buy(quantity, price float64) error {
	return errors.New("Not implemented")
}

// Sell -
func (m *gdax) Sell(quantity, price float64) error {
	return errors.New("Not implemented")
}

// GetBalance -
func (m *gdax) GetBalance() (assets float64, currency float64, err error) {
	return 0, 0, errors.New("Not implemented")
}

// Run -
func (m *gdax) Run() {
	m.Listen()
}

func (m *gdax) Backfill(end time.Time) error {
	uns := end.Format("2006-01-02 15:04:05")
	fmt.Printf("Backfilling %s.%s up to %s\n", Name, m.product, uns)
	// TODO Skip time spans we already have
	var trades []*market.Trade
	total := 0
	cur := m.client.ListTrades(m.product)
	for cur.HasMore {
		if err := cur.NextPage(&trades); err == nil {
			for _, t := range trades {
				t.Market = Name
				t.Product = m.product
				t.ID = fmt.Sprintf("%s.%s.%d", t.Market, t.Product, t.TradeID)
			}
			if err := m.persistence.PutTrade(trades...); err != nil {
				fmt.Println("Could not put trades", err)
				return err
			}
			total += len(trades)
			lt := trades[len(trades)-1]
			if lt.Time.Before(end) {
				fmt.Printf("Saved %d trades; Done!\n", total)
				return nil
			}
			fmt.Printf("Saved %d trades, %0.2f hours left.\n", total, lt.Time.Sub(end).Hours())
		}
	}
	return nil
}
