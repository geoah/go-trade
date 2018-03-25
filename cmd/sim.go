package cmd

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/spf13/cobra"

	agr "github.com/geoah/go-trade/aggregator"
	mrk "github.com/geoah/go-trade/market"
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
	strategy, err = simple.New(emaWindow)
	if err != nil {
		log.WithError(err).Fatalf("Could not setup strategy")
	}

	// setup fake market
	market, err = fake.New(persistence, gdax.Name, productName, simLast, simAssetCapital, simCurrencyCapital)
	log.
		WithField("balance-assets", simAssetCapital).
		WithField("balance-currency", simCurrencyCapital).
		Info("Started market")

	// setup aggregator
	aggregator, err := agr.NewTimeAggregator(15 * time.Minute)
	// aggregator, err := agr.NewVolumeAggregator(aggregationVolumeLimit)
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
		// WithField("aggregation-volume-limit", aggregationVolumeLimit).
		Infof("Started trading")

	// start market
	market.Run()

	// print data
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
	ioutil.WriteFile("data-sim.json", bs, 0644)

	log.Warnf("Completed simulation with %d actions", trader.Trades)

	// print data
	// data := make([][]interface{}, len(trader.Candles))
	// csv := "date,open,high,low,close\n"
	// for i, c := range trader.Candles {
	// 	data[i] = []interface{}{
	// 		c.Time.Format("2006-01-02 15:04:05"),
	// 		c.Low,
	// 		c.Open,
	// 		c.Close,
	// 		c.High,
	// 	}
	// 	csv += fmt.Sprintf("%s,%0.6f,%0.6f,%0.6f,%0.6f\n", data[i][0], c.Open, c.High, c.Low, c.Close)
	// }
	// bs, _ := json.Marshal(data)
	// ioutil.WriteFile("data-sim.json", bs, 0644)
	// ioutil.WriteFile("data-sim.csv", []byte(csv), 0644)
}
