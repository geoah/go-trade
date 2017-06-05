package gdax

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	logrus "github.com/Sirupsen/logrus"
	ws "github.com/gorilla/websocket"
	exchange "github.com/preichenberger/go-coinbase-exchange"

	market "github.com/geoah/go-trade/market"
	persistence "github.com/geoah/go-trade/persistence"
	utils "github.com/geoah/go-trade/utils"
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

	secret     string
	key        string
	passphrase string

	balanceCacheValid    bool
	balanceCacheAsset    float64
	balanceCacheCurrency float64
}

// New gdax market
func New(persistence persistence.Persistence, product string) (market.Market, error) {
	secret := os.Getenv("COINBASE_SECRET")
	key := os.Getenv("COINBASE_KEY")
	passphrase := os.Getenv("COINBASE_PASSPHRASE")

	// TODO Validate product for market

	mrk := &gdax{
		product:     strings.ToUpper(product),
		handlers:    []market.TradeHandler{},
		client:      exchange.NewClient(secret, key, passphrase),
		persistence: persistence,
		secret:      secret,
		key:         key,
		passphrase:  passphrase,
	}

	return mrk, nil
}

func (m *gdax) Listen() {
	var wsDialer ws.Dialer
	wsConn, _, err := wsDialer.Dial("wss://ws-feed.gdax.com", nil)
	if err != nil {
		println(err.Error())
	}

	time.Sleep(time.Second)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature, err := generateWebsocketSig(timestamp, m.secret)
	if err != nil {
		logrus.WithError(err).Fatalf("Could not subscribe to gdax ws")
	}
	subscribe := map[string]string{
		"type":       "subscribe",
		"product_id": m.product,
		"signature":  signature,
		"key":        m.key,
		"passphrase": m.passphrase,
		"timestamp":  timestamp,
	}
	if err := wsConn.WriteJSON(subscribe); err != nil {
		println("gdax ws sub error", err.Error())
	}

	message := exchange.Message{}
	for true {
		if err := wsConn.ReadJSON(&message); err != nil {
			logrus.WithError(err).Errorf("gdax ws read error")
			break
		}
		if message.Type == "match" {
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
					h.Handle(c) // TODO Handle error
				}
			}
		}
	}
}

func (m *gdax) asset() string {
	return strings.ToUpper(strings.Split(m.product, "-")[0])
}

func (m *gdax) currency() string {
	return strings.ToUpper(strings.Split(m.product, "-")[1])
}

// Notify -
func (m *gdax) Register(handler market.TradeHandler) {
	m.handlers = append(m.handlers, handler)
}

// Buy -
func (m *gdax) Buy(size, price float64) error {
	order := &exchange.Order{
		Price:       price,
		Size:        size,
		Side:        "buy",
		ProductId:   m.product,
		PostOnly:    true, // TODO Maker
		TimeInForce: "GTT",
		CancelAfter: "min",
	}
	_, err := m.client.CreateOrder(order)
	if err != nil {
		return err
	}
	m.balanceCacheValid = false
	return nil
}

// Sell -
func (m *gdax) Sell(size, price float64) error {
	order := &exchange.Order{
		Price:       price,
		Size:        size,
		Side:        "sell",
		ProductId:   m.product,
		PostOnly:    true, // TODO Maker
		TimeInForce: "GTT",
		CancelAfter: "min",
	}
	_, err := m.client.CreateOrder(order)
	if err != nil {
		return err
	}
	m.balanceCacheValid = false
	return nil
}

// GetBalance -
func (m *gdax) GetBalance() (assets float64, currency float64, err error) {
	if m.balanceCacheValid {
		return m.balanceCacheAsset, m.balanceCacheCurrency, nil
	}
	acs, err := m.client.GetAccounts()
	if err != nil {
		return 0, 0, err
	}
	cast := m.asset()
	ccur := m.currency()
	ast := 0.0
	cur := 0.0
	for _, acc := range acs {
		switch acc.Currency {
		case cast:
			ast = utils.TrimFloat64(acc.Available, 6)
		case ccur:
			cur = utils.TrimFloat64(acc.Available, 2)
		}
	}
	m.balanceCacheAsset = ast
	m.balanceCacheCurrency = cur
	return ast, cur, nil
}

// Run -
func (m *gdax) Run() {
	for {
		m.Listen()
	}
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

func generateWebsocketSig(timestamp, secret string) (string, error) {
	url := "/users/self"
	message := fmt.Sprintf("%sGET%s", timestamp, url)

	key, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", err
	}

	signature := hmac.New(sha256.New, key)
	_, err = signature.Write([]byte(message))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature.Sum(nil)), nil
}
