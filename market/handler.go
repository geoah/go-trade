package market

// CandleHandlerFunc -
type CandleHandlerFunc func(candle *Candle) error

// TradeHandlerFunc -
type TradeHandlerFunc func(trade *Trade) error

// CandleHandler -
type CandleHandler interface {
	HandleCandle(candle *Candle) error
}

// UpdateHandler -
type UpdateHandler interface {
	HandleUpdate(update *Update) error
}

// TradeHandler -
type TradeHandler interface {
	HandleTrade(trade *Trade) error
}
