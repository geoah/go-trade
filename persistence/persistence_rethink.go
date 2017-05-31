package persistence

import (
	"time"

	r "gopkg.in/gorethink/gorethink.v3"

	market "github.com/geoah/go-trade/market"
)

const (
	rethinkdbTradesTable     = "trades"
	rethinkdbTradesTimeIndex = "time"
)

// NewRethinkDB -
func NewRethinkDB(session *r.Session, database string) (Persistence, error) {
	return &rethinkdb{
		session:  session,
		database: database,
	}, nil
}

type rethinkdb struct {
	session  *r.Session
	database string
}

func (p *rethinkdb) PutTrade(trades ...*market.Trade) error {
	opts := r.InsertOpts{Conflict: "replace"}
	_, err := r.DB(p.database).
		Table(rethinkdbTradesTable).
		Insert(trades, opts).
		RunWrite(p.session)
	if err != nil {
		return err
	}
	return nil
}

func (p *rethinkdb) GetTrades(mrk, prd string, start, end time.Time) ([]*market.Trade, error) {
	cur, err := r.DB(p.database).Table(rethinkdbTradesTable).
		Between(start, end, r.BetweenOpts{
			Index:      rethinkdbTradesTimeIndex,
			RightBound: "closed",
		}).
		Filter(map[string]interface{}{
			"market":  mrk,
			"product": prd,
		}).
		OrderBy(r.Asc("time"), r.Asc("trade_id")).
		Run(p.session)
	if err != nil {
		return nil, err
	}
	trades := []*market.Trade{}
	if err := cur.All(&trades); err != nil {
		return nil, err
	}
	for _, trade := range trades {
		trade.Historic = true
	}
	return trades, nil
}
