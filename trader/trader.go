package trader

import (
	"github.com/Sirupsen/logrus"

	market "github.com/geoah/go-trade/market"
	strategy "github.com/geoah/go-trade/strategy"
	utils "github.com/geoah/go-trade/utils"
)

// Trader -
type Trader struct {
	strategy strategy.Strategy
	market   market.Market

	firstPriceSeen       float64
	firstAssetBalance    float64
	firstCurrencyBalance float64

	assetRounding    int
	currencyRounding int

	lastBuy  float64
	lastSell float64
	ready    bool

	Candles []*market.Candle
	Trades  int
}

// New trader
func New(market market.Market, strategy strategy.Strategy, assetRounding, currencyRounding int) (*Trader, error) {
	ast, cur, _ := market.GetBalance()
	return &Trader{
		strategy:             strategy,
		market:               market,
		firstAssetBalance:    ast,
		firstCurrencyBalance: cur,
		assetRounding:        assetRounding,
		currencyRounding:     currencyRounding,
	}, nil
}

// HandleUpdate -
func (t *Trader) HandleUpdate(update *market.Update) error {
	ast, cur, err := t.market.GetBalance()
	if err != nil {
		logrus.WithError(err).Warnf("Could not get balance")
		return nil
	}

	tlog := logrus.
		WithField("AST", ast).
		WithField("CUR", cur).
		WithField("Size", update.Size).
		WithField("Price", update.Price)

	switch update.Action {
	case market.Buy:
		tlog.Infof("Bought")
		t.lastBuy = update.Price
	case market.Sell:
		tlog.Errorf("Sold")
		t.lastSell = update.Price
	case market.Cancel:
		tlog.Warnf("Canceled")
	}

	return nil
}

// HandleCandle new candle
func (t *Trader) HandleCandle(candle *market.Candle) error {
	// logrus.WithField("candle", candle).Debug("Handling candle")
	// ready is a hack to make sure we don't sell insanely low when we start
	if t.ready == false {
		t.ready = true
		t.lastBuy = candle.Close
	}
	// TODO Move this and stream it
	t.Candles = append(t.Candles, candle)
	// TODO Lock simple? Not sure what for
	if t.firstPriceSeen == 0.0 {
		t.firstPriceSeen = candle.Close
	}
	action, err := t.strategy.HandleCandle(candle)
	if err != nil {
		logrus.WithError(err).Fatalf("Strategy could not handle trade")
	}
	// TODO random quantity to buy/sell is not clever, move to strategy
	qnt := 0.0
	switch action {
	case market.Hold:
		logrus.
			WithField("ACT", "Hold").
			Debugf("Strategy says")
		return nil
	case market.Buy:
		logrus.
			WithField("ACT", "Buy").
			Debugf("Strategy says")
		// act = "BUY"
		// get market price
		prc := candle.Close
		// figure how much can we buy
		_, cur, _ := t.market.GetBalance()
		// max assets we can buy
		// limit currency a bit
		// TODO Make configurable
		mas := cur / prc // * 0.5 // * 0.99
		// make sure we have enough currency to buy with
		if utils.TrimFloat64(mas, 5) == 0 {
			// nevermind
			logrus.
				WithField("CUR", cur).
				WithField("mas", mas).
				Debugf("Ignoring strategy, cannot buy, not enough money.")
			return nil
		}
		// random quantity of assets to buy
		qnt = t.quantity(mas)
		if qnt == 0.0 {
			logrus.
				WithField("CUR", cur).
				WithField("mas", mas).
				WithField("qnt", qnt).
				Debugf("Ignoring strategy, cannot buy, quantity too low.")
			return nil
		}
		prc = utils.TrimFloat64(prc, t.currencyRounding)
		atLeastPct := 1.001
		if prc >= t.lastSell/atLeastPct {
			logrus.WithField("last_sell", t.lastSell).
				WithField("prc", prc).
				Warnf("Ignoring strategy, I'm not buying, margin too small.")
			return nil
		}
		err = t.market.Buy(qnt, prc)
		if err != nil {
			logrus.WithError(err).Warnf("Could not buy assets")
			return nil
		}
		candle.Event = &market.Event{
			Action: string(market.Buy),
		}
		t.Trades++

	case market.Sell:
		logrus.
			WithField("ACT", "Sell").
			Debugf("Strategy says")
		// act = "SEL"
		// get market price
		prc := candle.Close
		// max assets we can sell
		// limit currency a bit
		// TODO Make configurable
		ast, _, _ := t.market.GetBalance()
		mas := ast * 0.99
		qnt = t.quantity(mas)
		if qnt == 0.0 {
			logrus.
				WithField("AST", ast).
				WithField("mas", mas).
				WithField("qnt", qnt).
				Debugf("Ignoring strategy, cannot sell, quantity too low.")
			return nil
		}
		prc = utils.TrimFloat64(prc, t.currencyRounding)
		atLeastPct := 1.005
		if t.lastBuy*atLeastPct >= prc {
			logrus.WithField("last_buy", t.lastBuy).
				WithField("prc", prc).
				Warnf("Ignoring strategy, I'm not selling, margin too small.")
			return nil
		}
		if t.lastBuy >= prc {
			logrus.WithField("last_buy", t.lastBuy).
				WithField("prc", prc).
				Warnf("Ignoring strategy, I'm not selling, selling price lower than what I bought at.")
			return nil
		}
		err = t.market.Sell(qnt, prc)
		if err != nil {
			logrus.
				WithError(err).
				WithField("AST", ast).
				Warnf("Could not sell assets")
			return nil
		}
		candle.Event = &market.Event{
			Action: string(market.Sell),
		}
		t.Trades++
	default:
		logrus.WithField("action", action).Fatalf("Strategy said something weird")
	}

	return nil
}

func (t *Trader) quantity(hardMax float64) float64 {
	hardMin := 0.01
	pct := 1.0 // 0.9

	// check if we have enough to sell
	if hardMax < hardMin {
		return 0
	}

	// reduce our quantity
	qnt := utils.TrimFloat64(hardMax*pct, t.assetRounding)

	// trim hardMax
	hardMax = utils.TrimFloat64(hardMax, t.assetRounding)

	// make sure we have enough to sell
	if qnt < hardMin {
		// if not, just sell it all
		return hardMax
	}

	// also check that the remaining qnt is above hard min
	if hardMax-qnt < hardMin {
		// if not, again just sell it all
		return hardMax
	}

	// else, return the reduced quantity
	return qnt
}
