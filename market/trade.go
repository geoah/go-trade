package market

import (
	"time"
)

// Trade -
type Trade struct {
	ID       string    `json:"-" gorethink:"id"`
	Market   string    `json:"-" gorethink:"market"`
	Product  string    `json:"-" gorethink:"product"`
	TradeID  int       `json:"trade_id,number" gorethink:"trade_id"`
	Price    float64   `json:"price,string" gorethink:"price"`
	Size     float64   `json:"size,string" gorethink:"size"`
	Time     time.Time `json:"time,string" gorethink:"time"`
	Side     string    `json:"side" gorethink:"side"`
	Historic bool      `json:"-" gorethink:"-"`
}
