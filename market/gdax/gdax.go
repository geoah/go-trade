package gdax

import (
	"errors"
	"os"

	ws "github.com/gorilla/websocket"
	exchange "github.com/preichenberger/go-coinbase-exchange"

	market "github.com/geoah/go-trade/market"
)

// gdax -
type gdax struct {
	product  string
	handlers []market.Handler
	client   *exchange.Client
}

// New gdax market
func New(product string) (market.Market, error) {
	secret := os.Getenv("COINBASE_SECRET")
	key := os.Getenv("COINBASE_KEY")
	passphrase := os.Getenv("COINBASE_PASSPHRASE")

	mrk := &gdax{
		product:  product,
		handlers: []market.Handler{},
		client:   exchange.NewClient(secret, key, passphrase),
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
		println(err.Error())
	}
	message := exchange.Message{}
	for true {
		if err := wsConn.ReadJSON(&message); err != nil {
			println(err.Error())
			break
		}
		if message.Type == "match" && message.Reason == "filled" {
			c := &market.Candle{
				Low:    message.Price,
				High:   message.Price,
				Volume: message.Size,
				Time:   message.Time.Time().UTC(),
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
func (m *gdax) Notify(handler market.Handler) {
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
	return 0, 0, nil
}

// Run -
func (m *gdax) Run() {
	m.Listen()
}
