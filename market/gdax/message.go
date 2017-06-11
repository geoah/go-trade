package gdax

import (
	exchange "github.com/preichenberger/go-coinbase-exchange"
)

// Message -
type Message struct {
	Type          string        `json:"type"`
	ProductID     string        `json:"product_id"`
	TradeID       int           `json:"trade_id,number"`
	OrderID       string        `json:"order_id"`
	Sequence      int           `json:"sequence,number"`
	MakerOrderID  string        `json:"maker_order_id"`
	TakerOrderID  string        `json:"taker_order_id"`
	Time          exchange.Time `json:"time,string"`
	RemainingSize float64       `json:"remaining_size,string"`
	NewSize       float64       `json:"new_size,string"`
	OldSize       float64       `json:"old_size,string"`
	Size          float64       `json:"size,string"`
	Price         float64       `json:"price,string"`
	Side          string        `json:"side"`
	Reason        string        `json:"reason"`
	OrderType     string        `json:"order_type"`
	Funds         float64       `json:"funds,string"`
	NewFunds      float64       `json:"new_funds,string"`
	OldFunds      float64       `json:"old_funds,string"`
	Message       string        `json:"message"`

	// private
	TakerUserID    string `json:"taker_user_id"`
	TakerProfileID string `json:"taker_profile_id"`
	UserID         string `json:"user_id"`
	ProfileID      string `json:"profile_id"`
	ClientOID      string `json:"client_oid"`
}
