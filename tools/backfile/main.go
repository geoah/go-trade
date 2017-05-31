package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	exchange "github.com/preichenberger/go-coinbase-exchange"
)

func gf(fp string) *os.File {
	f, err := os.OpenFile(fp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	return f
}

func main() {
	product := "ETH-BTC"
	secret := os.Getenv("COINBASE_SECRET")
	key := os.Getenv("COINBASE_KEY")
	passphrase := os.Getenv("COINBASE_PASSPHRASE")

	client := exchange.NewClient(secret, key, passphrase)

	diff := -1 * time.Minute
	end := time.Now().Add(time.Hour * -24).Round(time.Hour * 24)

	edt := end.UTC().Format("2006-01-02.json")
	ldt := edt
	f := gf(edt)

	for {
		edt = end.UTC().Format("2006-01-02.json")
		if ldt != edt {
			f.Close()
			f = gf(edt)
		}
		start := end.Add(diff)
		params := exchange.GetHistoricRatesParams{
			Start:       start,
			End:         end,
			Granularity: 1,
		}
		rates, err := client.GetHistoricRates(product, params)
		if err != nil {
			// fmt.Println("err", err)
			time.Sleep(250 * time.Millisecond)
			continue
		}
		for _, rate := range rates {
			// fmt.Println(trade.TradeId)
			bs, _ := json.Marshal(rate)
			// fmt.Println(string(bs))
			if _, err = f.WriteString(string(bs) + "\n"); err != nil {
				panic(err)
			}
		}

		fmt.Println(start.UTC().Format("2006-01-02T15:04:05-0700"), end.UTC().Format("2006-01-02T15:04:05-0700"), len(rates))

		end = start
		time.Sleep(275 * time.Millisecond)
	}
	// var trades []exchange.Trade
	// cursor := client.ListTrades("ETH-BTC")
	// for cursor.HasMore {
	// 	if err := cursor.NextPage(&trades); err != nil {
	// 		fmt.Println("err", err) // TODO Handle err
	// 	} else {
	// 		for _, trade := range trades {
	// 			// fmt.Println(trade.TradeId)
	// 			bs, _ := json.Marshal(trade)
	// 			fmt.Println(string(bs))
	// 		}
	// 	}
	// }

	// var wsDialer ws.Dialer
	// wsConn, _, err := wsDialer.Dial("wss://ws-feed.gdax.com", nil)
	// if err != nil {
	// 	println(err.Error())
	// }
	// subscribe := map[string]string{
	// 	"type":       "subscribe",
	// 	"product_id": "BTC-USD",
	// }
	// if err := wsConn.WriteJSON(subscribe); err != nil {
	// 	println(err.Error())
	// }
	// message := exchange.Message{}
	// for true {
	// 	if err := wsConn.ReadJSON(&message); err != nil {
	// 		println(err.Error())
	// 		break
	// 	}
	// 	if message.Type == "match" && message.Reason == "filled" {
	// 		bs, _ := json.Marshal(message)
	// 		fmt.Println(string(bs))
	// 		// println("Got a match", message.MakerOrderId, message.Price, message.Size)
	// 	}
	// }
}
