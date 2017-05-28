package market

// Market -
type Market interface {
	Notify(handler Handler)
	GetBalance() (assets float64, currency float64, err error)
	Buy(quantity, price float64) error
	Sell(quantity, price float64) error
	Run()
}
