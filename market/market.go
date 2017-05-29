package market

import (
	"time"
)

// Market -
type Market interface {
	Notify(handler TradeHandler)
	GetBalance() (assets float64, currency float64, err error)
	Buy(quantity, price float64) error
	Sell(quantity, price float64) error
	Run()
	Backfill(end time.Time) error
}
