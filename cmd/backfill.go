package cmd

import (
	"time"

	"github.com/geoah/go-trade/market/gdax"
	"github.com/spf13/cobra"
)

var (
	backfillDays = 1
)

// backfillCmd represents the backfill command
var backfillCmd = &cobra.Command{
	Use:   "backfill",
	Short: "Download historical trades for analysis and simulations",
	Run:   backfill,
}

func init() {
	RootCmd.AddCommand(backfillCmd)
	backfillCmd.PersistentFlags().IntVar(&backfillDays, "days", 1, "Number of days to backfill")
}

func backfill(cmd *cobra.Command, args []string) {
	var err error

	// setup gdax
	market, err = gdax.New(persistence, productName)
	if err != nil {
		log.WithError(err).Fatalf("Could not create market")
	}

	// backfill market
	market.Backfill(time.Now().Add(-24 * time.Duration(backfillDays) * time.Hour))
}
