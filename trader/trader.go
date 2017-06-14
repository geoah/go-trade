package trader

import (
	"fmt"
	"math/rand"

	logrus "github.com/Sirupsen/logrus"

	market "github.com/geoah/go-trade/market"
	strategy "github.com/geoah/go-trade/strategy"
	utils "github.com/geoah/go-trade/utils"
)

// Trader -
type Trader struct {
	strategy strategy.Strategy
	market   market.Market

	minSize          float64
	assetRounding    int
	currencyRounding int

	lastBuys  map[float64]float64 // map[round(price,2)]amount
	lastSells map[float64]float64 // map[round(price,2)]amount

	ready bool

	// Candles []*market.Candle
	Trades            int
	firstAssetBalance float64
	lastAssetBalance  float64
}

// New trader
func New(market market.Market, strategy strategy.Strategy, assetRounding, currencyRounding int, minSize float64) (*Trader, error) {
	return &Trader{
		strategy:         strategy,
		market:           market,
		minSize:          minSize,
		assetRounding:    assetRounding,
		currencyRounding: currencyRounding,
		lastBuys:         map[float64]float64{},
		lastSells:        map[float64]float64{},
	}, nil
}

// HandleUpdate -
func (t *Trader) HandleUpdate(update *market.Update) error {
	ast, cur, err := t.market.GetBalance()
	if err != nil {
		logrus.WithError(err).Warnf("Could not get balance")
		return nil
	}

	cbal := utils.TrimFloat64(ast+(cur/update.Price), t.assetRounding)
	dbal := utils.TrimFloat64(((cbal-t.firstAssetBalance)/t.firstAssetBalance)*100, 2)
	t.lastAssetBalance = cbal
	logrus.
		WithField("asset_balance", fmt.Sprintf("%0.4f", ast)).
		WithField("currency_balance", utils.TrimFloat64(cur, t.currencyRounding)).
		WithField("converted_asset_balance", fmt.Sprintf("%0.4f", cbal)).
		WithField("size", fmt.Sprintf("%0.4f", update.Size)).
		WithField("price", fmt.Sprintf("%0.2f", update.Price)).
		WithField("action", update.Action).
		WithField("inc", fmt.Sprintf("%0.2f%%", dbal)).
		Infof("Received update")

	switch update.Action {
	case market.Buy:
		// add to last buys
		tprc := utils.TrimFloat64(update.Price, t.currencyRounding)
		if lb, ok := t.lastBuys[tprc]; ok {
			lb += update.Size
		} else {
			t.lastBuys[tprc] = update.Size
		}

		// cleanup sells
		leftSize := update.Size
		totalSize := 0.0
		for lastPrice, lastSize := range t.lastSells {
			// we need to remove as much as we just bought
			// if we bought lower than this, so we can remove some of it
			if update.Price <= lastPrice {
				// when there is no more size left, stop
				if leftSize <= 0 {
					break
				}
				// if there is more to reduce
				if lastSize <= leftSize {
					// reduce leftSize
					leftSize -= lastSize
					// empty this price
					t.lastSells[lastPrice] = 0
				} else {
					// reduce this price
					t.lastSells[lastPrice] -= leftSize
					// empty leftSize
					leftSize = 0
				}
			}
			totalSize += lastSize
		}

		logrus.
			WithField("total_size", totalSize).
			WithField("left_size", leftSize).
			WithField("size", update.Size).
			Debugf("Cleaned up sales")

	case market.Sell:
		// add to last sells
		tprc := utils.TrimFloat64(update.Price, t.currencyRounding)
		if lb, ok := t.lastSells[tprc]; ok {
			lb += update.Size
		} else {
			t.lastSells[tprc] = update.Size
		}

		// cleanup buys
		leftSize := update.Size
		totalSize := 0.0
		totalSizeNew := 0.0
		for lastPrice, lastSize := range t.lastBuys {
			// we need to remove as much as we just sold
			// if we sold higher than this, we can remove some of it
			totalSize += lastSize
			logrus.
				WithField("last_price", lastPrice).
				WithField("price", update.Price).
				WithField("size", update.Size).
				WithField("size_left", leftSize).
				Debugf("Removing last buys")
			if update.Price >= lastPrice {

				// when there is no more size left, skip
				if leftSize > 0 {
					// if there is more to reduce
					if lastSize <= leftSize {
						// reduce leftSize
						leftSize -= lastSize
						// empty this price
						t.lastBuys[lastPrice] = 0
					} else {
						// reduce this price
						t.lastBuys[lastPrice] -= leftSize
						// empty leftSize
						leftSize = 0
					}
				}
			}
			totalSizeNew += t.lastBuys[lastPrice]
		}

		logrus.
			WithField("total_size", totalSize).
			WithField("total_size_new", totalSizeNew).
			WithField("left_size", leftSize).
			WithField("size", update.Size).
			Debugf("Cleaned up buys")
	}

	t.Trades++

	return nil
}

// HandleCandle new candle
func (t *Trader) HandleCandle(candle *market.Candle) error {
	// logrus.WithField("candle", candle).Debug("Handling candle")
	// ready is a hack to make sure we don't sell insanely low when we start
	if t.ready == false {
		ast, cur, err := t.market.GetBalance()
		if err != nil {
			logrus.
				WithError(err).
				Errorf("Could not get balance for ready")
			return nil
		}
		t.ready = true
		t.lastBuys[utils.TrimFloat64(candle.High, t.currencyRounding)] = ast
		t.lastSells[utils.TrimFloat64(candle.Low, t.currencyRounding)] = cur / candle.High
		t.firstAssetBalance = utils.TrimFloat64(ast+(cur/candle.High), t.assetRounding)

		logrus.
			WithField("lb", t.lastBuys).
			WithField("ls", t.lastSells).
			Warnf("lb/ls")
	}

	// ask strategy to tells us what to do
	action, err := t.strategy.HandleCandle(candle)
	if err != nil {
		logrus.WithError(err).Fatalf("Strategy could not handle trade")
	}

	switch action {
	case market.Hold:
		logrus.Debugf("Strategy says Hold")

	case market.Buy:
		logrus.Debugf("Strategy says Buy")

		// get market price
		price := utils.TrimFloat64(candle.Close, t.currencyRounding)

		// get our currency balance
		_, cur, _ := t.market.GetBalance()

		// figure out the max size we can actually buy
		maxSize := utils.TrimFloat64(cur/price, t.assetRounding)

		// figure out how much we have margin to buy
		size := t.checkMargin(action, price, maxSize)

		// limit size if we can't buy enough
		if size > maxSize {
			size = maxSize
		}

		// trim size so we don't spend everything at once
		size = t.quantity(size)

		// check against minimum size one last time
		if size < t.minSize {
			logrus.
				WithField("size", size).
				Debugf("Size below minimum")
			return nil
		}

		// submit order
		if err := t.market.Buy(size, price); err != nil {
			logrus.WithError(err).Warnf("Could not buy assets")
			return nil
		}

		logrus.
			WithField("action", action).
			WithField("price", price).
			WithField("size", size).
			Debugf("Submitted order")

	case market.Sell:
		logrus.Debugf("Strategy says Sell")

		// get market price
		price := candle.Close

		// get our asset balance
		ast, _, _ := t.market.GetBalance()

		// figure out what we have margin to buy
		size := t.checkMargin(action, price, ast)

		// trim size so we don't sell everything at once
		size = t.quantity(size)

		// check against minimum size one last time
		if size < t.minSize {
			logrus.
				WithField("size", size).
				Debugf("Size below minimum")
			return nil
		}

		// submit order
		if err := t.market.Sell(size, price); err != nil {
			logrus.
				WithError(err).
				WithField("asset_balance", ast).
				WithField("size", size).
				WithField("price", price).
				Warnf("Could not sell assets")
			return nil
		}

		logrus.
			WithField("action", action).
			WithField("price", price).
			WithField("size", size).
			Debugf("Submitted order")

	default:
		logrus.WithField("action", action).Fatalf("Strategy said something weird")
	}

	return nil
}

func (t *Trader) checkMargin(action market.Action, price, balance float64) (size float64) {
	// round price just in case
	price = utils.TrimFloat64(price, t.currencyRounding)

	// set profit margins
	buyPM := 1.001
	sellPM := 1.003

	total := 0.0

	switch action {
	case market.Buy:
		// find how much we can buy without making a loss
		for lastPrice, lastSize := range t.lastSells {
			// as long as the last price is lower than the current price
			if price <= lastPrice*buyPM {
				// we can buy this batch
				size += lastSize
			}
			// add total
			total += lastSize
		}

	case market.Sell:
		// find how much we can sell without making a loss
		for lastPrice, lastSize := range t.lastBuys {
			// as long as last price is greater than current price
			if price >= lastPrice*sellPM {
				// we can sell this batch
				size += lastSize
			}
			// add total
			total += lastSize
		}
	}

	// cleanup buys
	for lastPrice, lastSize := range t.lastBuys {
		if lastSize <= 0 {
			delete(t.lastBuys, lastPrice)
		}
	}

	// cleanup sells
	for lastPrice, lastSize := range t.lastSells {
		if lastSize <= 0 {
			delete(t.lastSells, lastPrice)
		}
	}

	// compare total size with balance and check if we have more
	// assets than we track
	if total < balance {
		// if so, add the difference to the proposed size
		diff := utils.TrimFloat64(balance-total, t.assetRounding)
		// size += diff
		logrus.
			WithField("diff", diff).
			Debugf("Adjusted margin size")
	}

	// round everything up
	size = utils.TrimFloat64(size, t.assetRounding)
	total = utils.TrimFloat64(total, t.assetRounding)

	logrus.
		WithField("action", action).
		WithField("total", total).
		WithField("size", size).
		WithField("price", price).
		Debugf("Margin result")

	return size
}

func (t *Trader) quantity(size float64) float64 {
	// make sure we have more than the minimum size
	if size < t.minSize {
		return 0
	}

	// trim size
	size = utils.TrimFloat64(size, t.assetRounding)

	// reduce our size
	rsize := utils.TrimFloat64(size*rand.Float64(), t.assetRounding)
	// rsize := size

	// check reduced size
	if rsize < t.minSize {
		// and if it's too low, fallback to the minimum size
		return t.minSize
	}

	// else return reduced size
	return rsize
}
