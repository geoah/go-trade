package market

import (
	"time"
)

// Candle -
type Candle struct {
	// Time is the start time
	Time time.Time `json:"time"`
	// Low is the lowest price
	Low float64 `json:"low"`
	// High is the highest price
	High float64 `json:"high"`
	// Open is the opening price (first trade)
	Open float64 `json:"open"`
	// Close is the closing price (last trade)
	Close float64 `json:"close"`
	// Volume of trading activity
	Volume float64 `json:"volume"`
	// Historic -
	Historic bool `json:"-"`

	Ema       float64 `json:"ema"`
	ChangePct float64 `json:"change_pct"`
	Event     *Event  `json:"event"`
}
