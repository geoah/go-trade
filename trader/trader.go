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

	lastBuy float64
	ready   bool

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

// Handle new candle
func (t *Trader) Handle(candle *market.Candle) error {
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
	action, err := t.strategy.Handle(candle)
	if err != nil {
		logrus.WithError(err).Fatalf("Strategy could not handle trade")
	}
	// TODO random quantity to buy/sell is not clever, move to strategy
	qnt := 0.0
	switch action {
	case strategy.Wait:
		logrus.
			WithField("ACT", "Hold").
			Debugf("Strategy says")
		return nil
	case strategy.Buy:
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
			return nil
		}
		// random quantity of assets to buy
		qnt = t.quantity(mas)
		if qnt == 0.0 {
			// logrus.Infof("Nil quantity")
			return nil
		}
		// adjust price to appear as a maker
		// TODO base this on the ema diff?
		// prc = prc / 1.0009
		prc = utils.TrimFloat64(prc, t.currencyRounding)
		// buy assets
		// logrus.
		// 	WithField("ACT", "Buy").
		// 	WithField("PRC", prc).
		// 	WithField("QNT", qnt).
		// 	Infof("Trying to")
		err = t.market.Buy(qnt, prc)
		if err != nil {
			logrus.WithError(err).Warnf("Could not buy assets")
			return nil
		}
		candle.Event = &market.Event{
			Action: string(strategy.Buy),
		}
		t.lastBuy = prc
		t.Trades++

		// log new balance
		ast, cur, err := t.market.GetBalance()
		if err != nil {
			logrus.WithError(err).Warnf("Could not get balance")
		}

		// initial assets
		iatl := t.firstAssetBalance + t.firstCurrencyBalance/t.firstPriceSeen
		// current assets
		catl := ast + cur/prc
		logrus.
			WithField("ACT", "BUY").
			WithField("PRC", utils.TrimFloat64(prc, 5)).
			WithField("AST%", utils.TrimFloat64(catl*100/iatl-100, 3)).
			Infof("Acted")

	case strategy.Sell:
		logrus.
			WithField("ACT", "Sell").
			Debugf("Strategy says")
		// act = "SEL"
		// get market price
		prc := candle.Close
		// if t.lastBuy > 0 && t.lastBuy >= prc {
		// 	logrus.WithField("last_buy", t.lastBuy).
		// 		WithField("prc", prc).
		// 		Errorf("DID NOT BUY, Fuck the strategy")
		// 	return nil
		// }
		// max assets we can sell
		// limit currency a bit
		// TODO Make configurable
		ast, _, _ := t.market.GetBalance()
		mas := ast * 0.99
		qnt = t.quantity(mas)
		if qnt == 0.0 {
			// logrus.Infof("Nil quantity")
			return nil
		}
		// adjust price to appear as a maker
		// TODO base this on the ema diff?
		// prc = prc * 1.0009
		prc = utils.TrimFloat64(prc, t.currencyRounding)
		// sell assets
		// logrus.
		// 	WithField("ACT", "Sell").
		// 	WithField("PRC", prc).
		// 	WithField("QNT", qnt).
		// 	Infof("Trying to")
		if t.lastBuy >= prc {
			logrus.WithField("last_buy", t.lastBuy).
				WithField("prc", prc).
				Warnf("Ignoring strategy, I'm not selling lower than I bought.")
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
			Action: string(strategy.Sell),
		}
		if t.lastBuy >= prc {
			logrus.
				WithField("last_buy", t.lastBuy).
				WithField("prc", prc).
				Warnf("WTF, Sold lower than bought...")
			// return nil
		}

		// log new balance
		ast, cur, err := t.market.GetBalance()
		if err != nil {
			logrus.WithError(err).Warnf("Could not get balance")
		}

		// initial assets
		iatl := t.firstAssetBalance + t.firstCurrencyBalance/t.firstPriceSeen
		// current assets
		catl := ast + cur/prc

		// t.lastBuy = 0
		logrus.
			WithField("lastBuy", t.lastBuy).
			WithField("ACT", "SEL").
			WithField("PRC", utils.TrimFloat64(prc, 5)).
			WithField("AST%", utils.TrimFloat64(catl*100/iatl-100, 3)).
			Errorf("Acted")
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
