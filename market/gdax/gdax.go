package gdax

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	uuid "github.com/google/uuid"
	ws "github.com/gorilla/websocket"
	exchange "github.com/preichenberger/go-coinbase-exchange"
	logrus "github.com/sirupsen/logrus"

	market "github.com/geoah/go-trade/market"
	persistence "github.com/geoah/go-trade/persistence"
	utils "github.com/geoah/go-trade/utils"
)

const (
	// Name of the market
	Name = "gdax"
)

var (
	ErrorOrderRejected = errors.New("Order rejected")
)

// gdax -
type gdax struct {
	product        string
	handlers       []market.TradeHandler
	updateHandlers []market.UpdateHandler
	client         *exchange.Client
	persistence    persistence.Persistence

	secret     string
	key        string
	passphrase string

	balanceCacheValid    bool
	balanceCacheAsset    float64
	balanceCacheCurrency float64

	profileID string
	clientOID string

	openOrders     map[string]*exchange.Order
	openOrdersLock sync.RWMutex
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
		openOrders:  map[string]*exchange.Order{},
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

	message := Message{}
	for {
		if err := wsConn.ReadJSON(&message); err != nil {
			logrus.WithError(err).Errorf("gdax ws read error")
			break
		}
		m.openOrdersLock.Lock()
		if message.Type == "error" {
			logrus.WithField("message", message).Errorf("GDAX Error")
		} else if _, ok := m.openOrders[message.OrderID]; ok {
			// this is our own order
			// if the order has been filled publish an update
			// TODO the type=match is not tested
			if message.Type == "match" {
				logrus.
					WithField("message", message).
					Warnf("Our order -- MATCHED")
				act := market.Sell
				if message.Side == "buy" {
					act = market.Buy
				}
				upd := &market.Update{
					Action: act,
					Price:  message.Price,
					Size:   message.Size,
					Time:   message.Time.Time(),
				}
				// logrus.WithField("update", upd).Warnf("Update on match")
				// TODO move to channels
				for _, h := range m.updateHandlers {
					if h != nil {
						h.HandleUpdate(upd) // TODO Handle error
					}
				}
			} else if message.Type == "done" {
				// if the order has been filled, publish an update
				if message.Reason == "filled" {
					act := market.Sell
					if message.Side == "buy" {
						act = market.Buy
					}
					upd := &market.Update{
						Action: act,
						Price:  message.Price,
						Size:   message.Size,
						Time:   message.Time.Time(),
					}
					// TODO move to channels
					for _, h := range m.updateHandlers {
						if h != nil {
							h.HandleUpdate(upd) // TODO Handle error
						}
					}
				} else {
					// report event
					upd := &market.Update{
						Action: market.Cancel,
						Price:  message.Price,
						Size:   message.Size,
						Time:   message.Time.Time(),
					}
					// TODO move to channels
					for _, h := range m.updateHandlers {
						if h != nil {
							h.HandleUpdate(upd) // TODO Handle error
						}
					}
				}
				// and remove from orders
				delete(m.openOrders, message.OrderID)
			}
		} else if message.ClientOID == m.clientOID {
			// our own orders
		} else if message.Type == "match" {
			t := &market.Trade{
				ID:      fmt.Sprintf("%s.%s.%d", Name, m.product, message.TradeID),
				Market:  Name,
				Product: m.product,
				TradeID: message.TradeID,
				Price:   message.Price,
				Size:    message.Size,
				Time:    message.Time.Time(),
				Side:    message.Side,
			}
			// TODO move to channels
			for _, h := range m.handlers {
				if h != nil {
					h.HandleTrade(t) // TODO Handle error
				}
			}
		}
		m.openOrdersLock.Unlock()
	}
}

func (m *gdax) asset() string {
	return strings.ToUpper(strings.Split(m.product, "-")[0])
}

func (m *gdax) currency() string {
	return strings.ToUpper(strings.Split(m.product, "-")[1])
}

// RegisterForTrades -
func (m *gdax) RegisterForTrades(handler market.TradeHandler) {
	m.handlers = append(m.handlers, handler)
}

// RegisterForUpdates -
func (m *gdax) RegisterForUpdates(handler market.UpdateHandler) {
	m.updateHandlers = append(m.updateHandlers, handler)
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
		ClientOID:   m.clientOID,
	}
	nord, err := m.client.CreateOrder(order)
	if err != nil {
		return err
	}
	if nord.Status == "rejected" {
		return ErrorOrderRejected
	}
	// add order to orders
	m.openOrdersLock.Lock()
	defer m.openOrdersLock.Unlock()
	m.openOrders[nord.Id] = &nord
	// report event
	logrus.
		WithField("price", utils.TrimFloat64(order.Price, 2)).
		WithField("size", utils.TrimFloat64(order.Size, 8)).
		Infof("Placed buy order")
	m.balanceCacheValid = false // TODO Remove balance cache
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
		ClientOID:   m.clientOID,
	}
	nord, err := m.client.CreateOrder(order)
	if err != nil {
		return err
	}
	if nord.Status == "rejected" {
		return ErrorOrderRejected
	}
	// add order to orders
	m.openOrdersLock.Lock()
	defer m.openOrdersLock.Unlock()
	m.openOrders[nord.Id] = &nord
	// report event
	logrus.
		WithField("price", utils.TrimFloat64(order.Price, 2)).
		WithField("size", utils.TrimFloat64(order.Size, 8)).
		Infof("Placed buy order")
	m.balanceCacheValid = false // TODO Remove balance cache
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
	// get profile
	// accs,err:=m.client.GetAccounts()
	// if err!=nil {
	// 	logrus.Warnf("Could not get account; will not be subscribing.")
	// } else {
	// 	m.profileID=accs[0].
	// }
	// new client id
	m.clientOID = uuid.New().String()
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
