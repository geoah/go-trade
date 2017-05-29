package cmd

import (
	gdax "github.com/geoah/go-trade/market/gdax"

	"github.com/geoah/go-trade/strategy/random"
	"github.com/spf13/cobra"
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
	strategy, err = random.New(0.7, 0.15, 0.15)
	if err != nil {
		log.WithError(err).Fatalf("Could not create strategy")
	}

	// setup gdax
	market, err = gdax.New(persistence, productName)
	if err != nil {
		log.WithError(err).Fatalf("Could not create market")
	}

	// attach handler
	market.Notify(handler)

	// start market
	market.Run()
}
