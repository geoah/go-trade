package trader

import (
	"github.com/Sirupsen/logrus"

	market "github.com/geoah/go-trade/market"
	strategy "github.com/geoah/go-trade/strategy"
	utils "github.com/geoah/go-trade/utils"
)

var (
	log = logrus.New()
)

// Trader -
type Trader struct {
	strategy strategy.Strategy
	market   market.Market

	firstPriceSeen       float64
	firstAssetBalance    float64
	firstCurrencyBalance float64
}

// New trader
func New(market market.Market, strategy strategy.Strategy) (*Trader, error) {
	ast, cur, _ := market.GetBalance()
	return &Trader{
		strategy:             strategy,
		market:               market,
		firstAssetBalance:    ast,
		firstCurrencyBalance: cur,
	}, nil
}

// Handle new candle
func (t *Trader) Handle(candle *market.Candle) error {
	// TODO Lock simple? Not sure what for
	if t.firstPriceSeen == 0.0 {
		t.firstPriceSeen = candle.Close
	}
	action, err := t.strategy.Handle(candle)
	if err != nil {
		log.WithError(err).Fatalf("Strategy could not handle trade")
	}
	act := "Holding"
	// TODO random quantity to buy/sell is not clever, move to strategy
	qnt := 0.0
	switch action {
	case strategy.Wait:
		log.Debugf("Strategy says wait")
		return nil
	case strategy.Buy:
		log.Debugf("Strategy says buy")
		act = "Bought"
		// get market price
		prc := candle.Close
		// figure how much can we buy
		_, cur, _ := t.market.GetBalance()
		// max assets we can buy
		mas := cur / prc
		// make sure we have enough currency to buy with
		if utils.RoundFloat64(mas, 5) == 0 {
			// nevermind
			return nil
		}
		// random quantity of assets to buy
		qnt = randomUpTo(mas)
		// round quantity
		if utils.RoundFloat64(qnt, 5) == utils.RoundFloat64(mas, 5) {
			qnt = mas
		}

		// buy assets
		log.Debugf("Trying to buy %0.2f at %0.2f", qnt, prc)
		err = t.market.Buy(qnt, prc)
		if err != nil {
			log.WithError(err).Warnf("Could not buy assets")
			return nil
		}
	case strategy.Sell:
		log.Debugf("Strategy says sell")
		act = "Sold"
		// get market price
		prc := candle.Close
		// max assets we can sell
		ast, _, _ := t.market.GetBalance()
		mas := ast
		// make sure we have enough assets to sell
		if utils.RoundFloat64(mas, 5) == 0 {
			// nevermind
			return nil
		} // random quantity of assets to sell
		qnt = randomUpTo(mas)
		// round qantity
		if utils.RoundFloat64(qnt, 5) == utils.RoundFloat64(mas, 5) {
			qnt = mas
		}

		// sell assets
		log.Debugf("Trying to sell %0.2f at %0.2f", qnt, prc)
		err = t.market.Sell(qnt, prc)
		if err != nil {
			log.WithError(err).Warnf("Could not sell assets")
			return nil
		}
	default:
		log.WithField("action", action).Fatalf("Strategy said something weird")
	}

	// log new balance
	ast, cur, err := t.market.GetBalance()
	if err != nil {
		log.WithError(err).Warnf("Could not get balance")
	}
	// current price
	prc := candle.Close
	// how much would our initial balance be worth
	itl := t.firstAssetBalance*t.firstPriceSeen + t.firstCurrencyBalance
	// total balance now
	ctl := ast*prc + cur
	// diff := ctl - itl

	// initial assets
	iatl := t.firstAssetBalance + t.firstCurrencyBalance/t.firstPriceSeen
	// current assets
	catl := ast + cur/prc

	log.
		WithField("ACT", act).
		WithField("QNT", utils.RoundFloat64(qnt, 5)).
		WithField("PRC", utils.RoundFloat64(prc, 5)).
		WithField("AST", utils.RoundFloat64(ast, 5)).
		WithField("CUR", utils.RoundFloat64(cur, 5)).
		WithField("CUR%", utils.RoundFloat64(ctl*100/itl-100, 3)).
		WithField("AST%", utils.RoundFloat64(catl*100/iatl-100, 3)).
		Infof(candle.Time.UTC().Format("2006-01-02 15:04:05"))
	return nil
}

// TODO this should not be random, maybe part of the strategy
func randomUpTo(max float64) float64 {
	return max
	// return (rand.Float64() * max)
}
