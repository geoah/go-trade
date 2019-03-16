# go-trade

## WARNING: DO NOT USE.

I would really suggest NOT using this for any kind of actual trading.
Even if the simulation tells you it will be profitable, it will not be.
The strategy was designed as a hack for relatively high volume and a bit or randomness, but it's too slow for it to actually be useful.

## Setup your env

### GDAX

Create an [API key](https://www.gdax.com/settings/api) and set the following env vars:

  * `COINBASE_SECRET`
  * `COINBASE_KEY`
  * `COINBASE_PASSPHRASE`

## Installation on OSX

* Install [golang](https://golang.org/doc/install) >= 1.7
* Install [golang dep](https://github.com/golang/dep).

* `dep ensure` to get all dependencies.
* `docker-compose up -d` to start rethinkdb in the background.
* `go run *.go backfill --product=ETH-USD --days=2` to get 2 days of `gdax.ETH-USD` historic trade data.
* `go run *.go sim --product=ETH-USD --last=2h --asset_capital=10 --currency_capital=1000` to simulate the the random strategy on the last day of `gdax.ETH-USD` trades.
* `go run *.go trade` to run the random strategy on realtime `gdax` trades.
