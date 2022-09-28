package models

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gomodule/redigo/redis"
	"github.com/shopspring/decimal"

	"quote/utils"
)

const RedisNotify = "quote"

type Quote struct {
	CommonModel
	Type          string          `json:"-"`
	BaseId        int             `json:"-"`
	QuoteId       int             `json:"-"`
	MarketId      int             `json:"-"`
	Timestamp     int64           `json:"timestamp"`
	Price         decimal.Decimal `json:"price" gorm:"type:decimal(32,16);default:0;"`
	Source        string          `json:"source"`
	Base          string          `sql:"-" json:"base"`
	Quote         string          `sql:"-" json:"quote"`
	Market        string          `sql:"-" json:"market"`
	QuoteCurrency Currency        `sql:"-" json:"-"`
}
type SimpleQuote struct {
	Timestamp int64           `json:"timestamp"`
	Price     decimal.Decimal `json:"price" `
	Source    string          `json:"source"`
	Base      string          `json:"base"`
	Quote     string          `json:"quote"`
	Market    string          `json:"market"`
}

func (quote *Quote) AfterFind() {
	quote.SetAttrs()
}
func (quote *Quote) SetAttrs() {
	quote.Source = strings.ToLower(strings.Split(quote.Type, "Quotes::")[1])
	for _, currency := range AllCurrencies {
		if currency.Id == quote.BaseId {
			quote.Base = currency.Symbol
		}
		if currency.Id == quote.QuoteId {
			quote.QuoteCurrency = currency
			quote.Quote = currency.Symbol
		}
	}
	if quote.MarketId != 0 {
		for _, market := range AllMarkets {
			if quote.MarketId == market.Id {
				quote.Market = market.Symbol
			}
		}
	}
}

func (quote *Quote) RedisKey() string {
	return fmt.Sprintf("Quotes:%v:%v:%v", quote.MarketId, quote.BaseId, quote.QuoteId)
}

func (quote *Quote) SaveToRedis() {
	dataRedis := utils.GetRedisConn("data")
	defer dataRedis.Close()
	b, err := json.Marshal(*quote)
	if err != nil {
		log.Println("error:", err)
	}
	dataRedis.Do("SETEX", quote.RedisKey(), 60, b)
}

func (quote *Quote) IsAnchored() (no bool) {
	quote.SetAttrs()
	for _, c := range []string{"usdt", "cnst"} {
		if quote.Quote == c {
			return true
		}
	}
	return
}

func (quote *Quote) IsLegal() (no bool) {
	quote.SetAttrs()
	for _, c := range []string{"usd", "cny"} {
		if quote.Quote == c {
			return true
		}
	}
	return
}
func (quote *Quote) AlreadyHave() (yes bool) {
	dataRedis := utils.GetRedisConn("data")
	defer dataRedis.Close()
	str, _ := redis.Bytes(dataRedis.Do("GET", quote.RedisKey()))
	var q Quote
	json.Unmarshal(str, &q)
	if quote.Timestamp < q.Timestamp {
		yes = true
	}
	return
}

func (quote *Quote) NotifyQuote() {
	b, err := json.Marshal(SimpleQuote{
		Timestamp: quote.Timestamp,
		Price:     quote.Price,
		Source:    quote.Source,
		Base:      quote.Base,
		Quote:     quote.Quote,
		Market:    quote.Market,
	})
	if err != nil {
		log.Println("error:", err)
	}
	utils.PublishToPubSubChannels(RedisNotify, &b)
}
