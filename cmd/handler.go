package cmd

import (
	"math/rand"

	"strings"

	mrk "github.com/geoah/go-trade/market"
	str "github.com/geoah/go-trade/strategy"
)

func handler(trade *mrk.Trade) error {
	action, err := strategy.Handle(trade)
	if err != nil {
		log.WithError(err).Fatalf("Strategy could not handle trade")
	}
	act := "Holding"
	// TODO random quantity to buy/sell is not clever, move to strategy
	qnt := 0.0
	switch action {
	case str.Wait:
		log.Debugf("Strategy says wait")
	case str.Buy:
		log.Debugf("Strategy says buy")
		act = "Buying"
		// get market price
		prc := trade.Price
		// figure how much can we buy
		_, cur, _ := market.GetBalance()
		// max assets we can buy
		mas := cur / prc
		// TODO make sure we have enough currency to buy with mas > 0
		// random quantity of assets to buy
		qnt = randomUpTo(mas)
		// buy assets
		log.Debugf("Trying to buy %0.2f at %0.2f", qnt, prc)
		if err := market.Buy(qnt, prc); err != nil {
			log.WithError(err).Warnf("Could not buy assets")
			break
		}
	case str.Sell:
		log.Debugf("Strategy says sell")
		act = "Selling"
		// get market price
		prc := trade.Price
		// max assets we can sell
		mas, _, _ := market.GetBalance()
		// TODO make sure we have enough assets to sell mas > 0
		// random quantity of assets to sell
		qnt = randomUpTo(mas)
		// sell assets
		log.Debugf("Trying to sell %0.2f at %0.2f", qnt, prc)
		if err := market.Sell(qnt, prc); err != nil {
			log.WithError(err).Warnf("Could not sell assets")
			break
		}
	default:
		log.WithField("action", action).Fatalf("Strategy said something weird")
	}

	// log new balance
	ast, cur, err := market.GetBalance()
	if err != nil {
		// TODO remove this, quick hack to not show negative diff
		if strings.Contains(err.Error(), "Not implemented") {
			ast = simAssetCapital
			cur = simCurrencyCapital
		}
	}
	// current price
	prc := trade.Price
	// how much would our initial balance be worth
	// itl := simAssetCapital*prc + simCurrencyCapital
	// total balance now
	// ctl := ast*prc + cur
	// diff := fmt.Sprintf("%0.4f", ctl-itl)
	log.
		WithField("ACT", act).
		WithField("QNT", qnt).
		WithField("PRC", prc).
		WithField("AST", ast).
		WithField("CUR", cur).
		// WithField("DIFF", diff).
		Infof(trade.Time.Format("2006-01-02 15:04:05") + " - " + trade.ID)
	return nil
}

// TODO this should not be random, maybe part of the strategy
func randomUpTo(max float64) float64 {
	return (rand.Float64() * max)
}
