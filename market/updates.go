package market

import "time"

// Update -
type Update struct {
	Action Action
	Price  float64
	Size   float64
	Time   time.Time
}
