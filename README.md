# go-trade

## Installation

* `glide up` to get all dependencies.
* `docker-compose up -d` to start rethinkdb in the background.
* `go install` to install `go-trade` binary.
* `go-trade backfill --product=ETH-USD --days=2` to get 2 days of `gdax.ETH-USD` historic trade data.
* `go-trade sim --product=ETH-USD --last=2h --asset_capital=10 --currency_capital=1000` to simulate the the random strategy on the last day of `gdax.ETH-USD` trades.
* `go-trade trade` to run the random strategy on realtive `gdax` trades.
