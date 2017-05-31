package cmd

import (
	"github.com/spf13/cobra"

	agr "github.com/geoah/go-trade/aggregator"
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
	// aggregator, err := agr.NewTimeAggregator(1 * time.Minute)
	aggregator, err := agr.NewVolumeAggregator(aggregationVolumeLimit)
	if err != nil {
		log.WithError(err).Fatalf("Could not setup aggregator")
	}

	// setup trader
	trader, err = trd.New(market, strategy)
	if err != nil {
		log.WithError(err).Fatalf("Could not setup trader")
	}

	// attach handlers
	market.Register(aggregator)
	aggregator.Register(trader)

	log.
		WithField("ema-window", emaWindow).
		WithField("aggregation-volume-limit", aggregationVolumeLimit).
		Infof("Started trading")

	// start market
	market.Run()
}
