package market

import "time"

// Candle -
type Candle struct {
	Time   time.Time `json:"Time"`
	Low    float64   `json:"Low"`
	High   float64   `json:"High"`
	Open   float64   `json:"Open"`
	Close  float64   `json:"Close"`
	Volume float64   `json:"Volume"`
}
