package cmd

import (
	"fmt"
	"os"

	logrus "github.com/Sirupsen/logrus"
	homedir "github.com/mitchellh/go-homedir"
	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	r "gopkg.in/gorethink/gorethink.v3"

	mrk "github.com/geoah/go-trade/market"
	per "github.com/geoah/go-trade/persistence"
	str "github.com/geoah/go-trade/strategy"
	trd "github.com/geoah/go-trade/trader"
)

var (
	log     *logrus.Logger
	cfgFile string

	marketName             string
	productName            string
	logLevel               string
	emaWindow              float64
	aggregationVolumeLimit float64

	persistence per.Persistence
	market      mrk.Market
	strategy    str.Strategy
	trader      *trd.Trader
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "go-trade",
	Short: "A proof of concept crypto trading bot",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.go-trade.yaml)")
	RootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level [debug/info/warn/error")

	RootCmd.PersistentFlags().StringVar(&productName, "product", "BTC-USD", "product name")
	RootCmd.PersistentFlags().Float64Var(&emaWindow, "ema-window", 100, "EMA window")
	RootCmd.PersistentFlags().Float64Var(&aggregationVolumeLimit, "aggregation-volume", 0.5, "Volume aggregation")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".go-trade" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".go-trade")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	setup()
}

func setup() {
	logrus.SetFormatter(&prefixed.TextFormatter{
		FullTimestamp:    true,
		QuoteEmptyFields: false,
		QuoteCharacter:   "",
	})

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	log = logrus.New()

	rs, err := r.Connect(r.ConnectOpts{
		Address: "localhost",
	})
	if err != nil {
		log.WithError(err).Fatalf("Could not connect to rethinkdb")
	}

	rDB := "trade"

	// TODO Create db, tables, and indexes only if they don't exist
	if _, err := r.DBCreate(rDB).RunWrite(rs); err != nil {
		// log.WithError(err).Fatalf("Could not create rethinkdb database")
	}
	if err := r.DB(rDB).TableCreate("trades").Exec(rs); err != nil {
		// log.WithError(err).Fatalf("Could not create rethinkdb table")
	}
	if err := r.DB(rDB).Table("trades").IndexCreate("time").Exec(rs); err != nil {
		// log.WithError(err).Fatalf("Could not create rethinkdb index")
	}
	if err := r.DB(rDB).Table("trades").IndexWait().Exec(rs); err != nil {
		log.WithError(err).Fatalf("Could not wait for indexes")
	}

	persistence, err = per.NewRethinkDB(rs, rDB)
	if err != nil {
		log.WithError(err).Fatalf("Could not create rethinkdb persistence")
	}

}
