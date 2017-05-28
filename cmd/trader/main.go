package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	logrus "github.com/sirupsen/logrus"

	market "github.com/geoah/go-trade/market"
	"github.com/geoah/go-trade/market/gdax"
	strategy "github.com/geoah/go-trade/strategy"
	random "github.com/geoah/go-trade/strategy/random"
)

var (
	// setup logging
	log = logrus.New()

	// dummy balance
	// assetBalance    = 0.0
	// currencyBalance = 1000.0

	// setup market
	mrk market.Market

	// setup strategy
	str strategy.Strategy
)

func main() {
	// seed random
	rand.Seed(time.Now().UTC().UnixNano())

	// setup logger
	logrus.SetLevel(logrus.DebugLevel)

	// load gdax historic candles
	// cs, err := getCandlesFromDir("./data/gdax-btc-usd-small/")
	// if err != nil {
	// 	log.WithError(err).Fatalf("Could not list files")
	// }

	// create new fake market
	// mrk, err = fake.New(cs, assetBalance, currencyBalance)
	var err error
	mrk, err = gdax.New("BTC-USD")
	if err != nil {
		log.WithError(err).Fatalf("Could not create market")
	}

	// setup strategy
	str, err = random.New(0.7, 0.15, 0.15)
	if err != nil {
		log.WithError(err).Fatalf("Could not create strategy")
	}

	// attach handler
	mrk.Notify(handler)

	// start market
	mrk.Run()
}

func handler(candle *market.Candle) error {
	action, err := str.Handle(candle)
	if err != nil {
		log.WithError(err).Fatalf("Strategy could not handle candle")
	}
	switch action {
	case strategy.Wait:
		log.Debugf("Strategy says wait")
		// return and don't log new balance
		return nil
	case strategy.Buy:
		log.Debugf("Strategy says buy")
		// get market price
		prc := candle.Low
		// figure how much can we buy
		_, cur, _ := mrk.GetBalance()
		// max assets we can buy
		mas := cur / prc
		// random quantify of assets to buy
		qnt := randomUpTo(mas)
		// buy assets
		log.Debugf("Trying to buy %0.2f at %0.2f", qnt, prc)
		if err := mrk.Buy(qnt, prc); err != nil {
			log.WithError(err).Warnf("Could not buy assets")
			break
		}
	case strategy.Sell:
		log.Debugf("Strategy says sell")
		// get market price
		prc := candle.High
		// max assets we can sell
		mas, _, _ := mrk.GetBalance()
		// random quantify of assets to buy
		qnt := randomUpTo(mas)
		// sell assets
		log.Debugf("Trying to sell %0.2f at %0.2f", qnt, prc)
		if err := mrk.Sell(qnt, prc); err != nil {
			log.WithError(err).Warnf("Could not sell assets")
			break
		}
	default:
		log.WithField("action", action).Fatalf("Strategy said something weird")
	}

	// log new balance
	ast, cur, _ := mrk.GetBalance()
	// current price
	// prc := candle.High
	// how much would our initial balance be worth
	// itl := assetBalance*prc + currencyBalance
	// total balance now
	// ctl := ast*prc + cur
	// diff := fmt.Sprintf("%0.4f", ctl-itl)
	log.
		WithField("AST", ast).
		WithField("CUR", cur).
		// WithField("DIFF", diff).
		Infof("New balance")
	return nil
}

func getCandlesFromDir(path string) ([]*market.Candle, error) {
	fs, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	cs := []*market.Candle{}
	dc := 0
	for _, f := range fs {
		if f.IsDir() {
			log.Warnf("We don't support nested data dirs yet")
			continue
		}
		if !strings.Contains(f.Name(), ".json") {
			log.Warnf("We don't support non JSON files yet")
			continue
		}
		dc++
		fpath := filepath.Join(path, f.Name())
		fcs, err := getCandlesFromFile(fpath)
		if err != nil {
			log.WithError(err).Errorf("Could not read file %s", fpath)
			return nil, err
		}
		cs = append(cs, fcs...)
	}
	log.Infof("> Loaded %d candles from directory %s from %d files", len(cs), path, dc)
	return cs, nil
}

func getCandlesFromFile(path string) ([]*market.Candle, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	cs := []*market.Candle{}
	d := json.NewDecoder(f)
	for {
		c := &market.Candle{}
		if err := d.Decode(&c); err == io.EOF {
			break
		} else if err != nil {
			return nil, err // TODO or continue
		}
		cs = append(cs, c)
	}
	log.Infof(">> Loaded %d candles from file %s", len(cs), path)
	return cs, nil
}

func randomUpTo(max float64) float64 {
	return (rand.Float64() * max)
}
