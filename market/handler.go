package market

// CandleHandler -
type CandleHandler func(candle *Candle) error

// TradeHandler -
type TradeHandler func(trade *Trade) error
