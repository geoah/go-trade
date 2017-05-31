package market

// CandleHandlerFunc -
type CandleHandlerFunc func(candle *Candle) error

// TradeHandlerFunc -
type TradeHandlerFunc func(trade *Trade) error

// CandleHandler -
type CandleHandler interface {
	Handle(candle *Candle) error
}

// TradeHandler -
type TradeHandler interface {
	Handle(trade *Trade) error
}
