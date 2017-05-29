package cmd

import (
	"time"

	"github.com/geoah/go-trade/market/fake"
	"github.com/geoah/go-trade/market/gdax"
	"github.com/geoah/go-trade/strategy/random"
	"github.com/spf13/cobra"
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
	strategy, err = random.New(0.7, 0.15, 0.15)
	if err != nil {
		log.WithError(err).Fatalf("Could not create strategy")
	}

	// setup fake market
	market, err = fake.New(persistence, gdax.Name, productName, simLast, simAssetCapital, simCurrencyCapital)

	// attach handler
	market.Notify(handler)

	// start market
	market.Run()
}
