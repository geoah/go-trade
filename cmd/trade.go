package cmd

import (
	"time"

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
	strategy, err = simple.New(20)
	if err != nil {
		log.WithError(err).Fatalf("Could not setup strategy")
	}

	// setup gdax
	market, err = gdax.New(persistence, productName)
	if err != nil {
		log.WithError(err).Fatalf("Could not create market")
	}

	// setup aggregator
	aggregator, err := agr.NewTimeAggregator(5 * time.Minute)
	// aggregator, err := agr.NewVolumeAggregator(100)
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

	// start market
	market.Run()
}
