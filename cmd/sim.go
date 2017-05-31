package cmd

import (
	"time"

	"github.com/spf13/cobra"

	agr "github.com/geoah/go-trade/aggregator"
	fake "github.com/geoah/go-trade/market/fake"
	gdax "github.com/geoah/go-trade/market/gdax"
	simple "github.com/geoah/go-trade/strategy/simple"
	trd "github.com/geoah/go-trade/trader"
)

var (
	simAssetCapital    = 0.0
	simCurrencyCapital = 1000.0
	simLast            = time.Hour
)

// simCmd represents the sim command
var simCmd = &cobra.Command{
	Use:   "sim",
	Short: "Run a simulation on backfilled data",
	Run:   sim,
}

func init() {
	RootCmd.AddCommand(simCmd)
	simCmd.Flags().Float64Var(&simAssetCapital, "asset_capital", 0.0, "Amount of start capital in asset")
	simCmd.Flags().Float64Var(&simCurrencyCapital, "currency_capital", 1000.0, "Amount of start capital in currency")
	simCmd.Flags().DurationVar(&simLast, "last", time.Hour, "Simulate the last hours/days/etc to sim. eg 1h")
}

func sim(cmd *cobra.Command, args []string) {
	var err error

	// setup strategy
	strategy, err = simple.New(2000)
	if err != nil {
		log.WithError(err).Fatalf("Could not setup strategy")
	}

	// setup fake market
	market, err = fake.New(persistence, gdax.Name, productName, simLast, simAssetCapital, simCurrencyCapital)
	log.
		WithField("AST", simAssetCapital).
		WithField("CUR", simCurrencyCapital).
		Info("Started market")

	// setup aggregator
	// aggregator, err := agr.NewTimeAggregator(1 * time.Minute)
	aggregator, err := agr.NewVolumeAggregator(100)
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
