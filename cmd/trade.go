package cmd

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/signal"
	"time"

	cobra "github.com/spf13/cobra"

	agr "github.com/geoah/go-trade/aggregator"
	mrk "github.com/geoah/go-trade/market"
	gdax "github.com/geoah/go-trade/market/gdax"
	simple "github.com/geoah/go-trade/strategy/simple"
	trd "github.com/geoah/go-trade/trader"
)

// tradeCmd represents the trade command
var tradeCmd = &cobra.Command{
	Use:   "trade",
	Short: "Run trading bot against live market data",
	Run:   trade,
}

func init() {
	RootCmd.AddCommand(tradeCmd)
	// tradeCmd.PersistentFlags().String("foo", "", "A help for foo")
	// tradeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func trade(cmd *cobra.Command, args []string) {
	var err error

	// setup strategy
	strategy, err = simple.New(emaWindow)
	if err != nil {
		log.WithError(err).Fatalf("Could not setup strategy")
	}

	// setup fake market
	market, err = gdax.New(persistence, productName)
	if ast, cur, err := market.GetBalance(); err != nil {
		log.WithError(err).Fatalf("Could not get first time balance")
	} else {
		log.
			WithField("balance-assets", ast).
			WithField("balance-currency", cur).
			Info("Started market")
	}

	// setup aggregator
	// aggregator, err := agr.NewTimeAggregator(30 * time.Second)
	aggregator, err := agr.NewVolumeAggregator(aggregationVolumeLimit)
	if err != nil {
		log.WithError(err).Fatalf("Could not setup aggregator")
	}

	// setup trader
	trader, err = trd.New(market, strategy, 8, 2) // TODO Get precision from market
	if err != nil {
		log.WithError(err).Fatalf("Could not setup trader")
	}

	// attach handlers
	market.RegisterForTrades(aggregator)
	market.RegisterForUpdates(trader)
	aggregator.Register(trader)

	log.
		WithField("ema-window", emaWindow).
		WithField("aggregation-volume-limit", aggregationVolumeLimit).
		Infof("Started trading")

	started := time.Now().UTC()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			log.WithField("sig", sig).Infof("Interrupted")
			now := time.Now()
			data := []*mrk.Candle{}
			for i, c := range trader.Candles {
				if c.Ema > 0 && i > 10 {
					c.Time = now
					now = now.Add(5 * time.Minute)
					data = append(data, c)
				}
			}
			bs, _ := json.Marshal(data)
			ioutil.WriteFile("data-trade-"+started.Format("2006-01-02T15:04:05Z")+".json", bs, 0644)
			log.Warnf("Completed trade with %d actions", trader.Trades)
			os.Exit(0)
		}
	}()

	// start market
	market.Run()
}
