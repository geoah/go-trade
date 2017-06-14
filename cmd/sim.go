package cmd

import (
	"time"

	logrus "github.com/Sirupsen/logrus"
	cobra "github.com/spf13/cobra"

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
	strategy, err = simple.New(emaWindow)
	if err != nil {
		logrus.WithError(err).Fatalf("Could not setup strategy")
	}

	// setup fake market
	start, err := time.Parse(time.RFC3339, "2017-06-08T00:00:00+00:00")
	if err != nil {
		logrus.WithError(err).Fatalf("Could not parse start time")
	}
	end, err := time.Parse(time.RFC3339, "2017-06-12T00:00:00+00:00")
	if err != nil {
		logrus.WithError(err).Fatalf("Could not parse end time")
	}
	market, err = fake.New(persistence, gdax.Name, productName, start, end, simAssetCapital, simCurrencyCapital)
	logrus.
		WithField("balance-assets", simAssetCapital).
		WithField("balance-currency", simCurrencyCapital).
		Info("Started market")

	// setup aggregator
	// aggregator, err := agr.NewTimeAggregator(1 * time.Minute)
	aggregator, err := agr.NewVolumeAggregator(aggregationVolumeLimit)
	if err != nil {
		logrus.WithError(err).Fatalf("Could not setup aggregator")
	}

	// setup trader
	trader, err = trd.New(market, strategy, 8, 2, 0.01) // TODO Get precision from market
	if err != nil {
		logrus.WithError(err).Fatalf("Could not setup trader")
	}

	// attach handlers
	market.RegisterForTrades(aggregator)
	market.RegisterForUpdates(trader)
	aggregator.Register(trader)

	logrus.
		WithField("ema-window", emaWindow).
		WithField("aggregation-volume-limit", aggregationVolumeLimit).
		Infof("Started trading")

	// start market
	market.Run()

	// print data
	// now := time.Now()
	// data := []*mrk.Candle{}
	// for _, c := range trader.Candles {
	// 	// if c.Ema > 0 && i > 10 {
	// 	// c.Time = now
	// 	// now = now.Add(5 * time.Minute)
	// 	data = append(data, c)
	// 	// }
	// }
	// bs, _ := json.Marshal(data)
	// ioutil.WriteFile("data-sim.json", bs, 0644)

	logrus.Warnf("Completed simulation with %d actions", trader.Trades)

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
