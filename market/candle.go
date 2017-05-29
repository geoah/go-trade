package market

import "time"

// Candle -
type Candle struct {
	// Time is the start time
	Time time.Time `json:"Time"`
	// Low is the lowest price
	Low float64 `json:"Low"`
	// High is the highest price
	High float64 `json:"High"`
	// Open is the opening price (first trade)
	Open float64 `json:"Open"`
	// Close is the closing price (last trade)
	Close float64 `json:"Close"`
	// Volume of trading activity
	Volume float64 `json:"Volume"`
}
