package redis

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/streadway/amqp"

	"quote/config"
	"quote/initializers"
	. "quote/models"
	"quote/utils"
)

func Save(market *Market, k *KLine) {
	dataRedis := utils.GetRedisConn("data")
	defer dataRedis.Close()

	log.Println((*market).Source, (*market).Symbol, (*k).Timestamp, (*k).Close)
	dataRedis.Send("zremrangebyscore", (*market).TimeLine(), (*k).Timestamp, (*k).Timestamp)
	dataRedis.Do("zadd", (*market).TimeLine(), (*k).Timestamp, (*k).Close)

	timestamp := (*k).Timestamp / 60000 * 60
	b, err := json.Marshal(KLine{
		Timestamp: timestamp,
		Open:      (*k).Open,
		Close:     (*k).Close,
		Low:       (*k).Low,
		High:      (*k).High,
		Vol:       (*k).Vol,
	})
	if err != nil {
		log.Println("error:", err)
	}
	dataRedis.Send("zremrangebyscore", (*market).KLine(1), timestamp, timestamp)
	dataRedis.Do("zadd", (*market).KLine(1), timestamp, b)

	db := utils.DbBegin()
	defer db.DbRollback()
	var quote Quote
	db.FirstOrInit(&quote, map[string]interface{}{
		"market_id": (*market).Id,
		"type":      "Quotes::" + strings.Title((*market).Source),
		"source":    (*market).Source,
		"base_id":   (*market).BaseId,
		"quote_id":  (*market).QuoteId,
	})
	quote.Timestamp = (*k).Timestamp
	quote.Price = (*k).Close
	db.Save(&quote)
	db.DbCommit()
	buildOtherKLines(dataRedis, market)
	createSubQuote(&quote)

	n, err := json.Marshal(NotifyKLine{
		KLine: KLine{
			Timestamp: timestamp,
			Open:      (*k).Open,
			Close:     (*k).Close,
			Low:       (*k).Low,
			High:      (*k).High,
			Vol:       (*k).Vol,
		},
		MarketId: (*market).Id,
		Period:   1,
	})
	if err != nil {
		log.Println("error:", err)
	}
	go func() {
		err = config.RabbitMqConnect.PublishMessageWithRouteKey(config.AmqpGlobalConfig.Exchange["fanout"]["k"], "#", "text/plain", false, false, &n, amqp.Table{}, amqp.Transient, "")
		if err != nil {
			log.Println("{error:", err, "}")
		}
	}()
	ticker := refreshTicker(dataRedis, market)
	t, err := json.Marshal(ticker)
	if err != nil {
		log.Println("error:", err)
	}
	go func() {
		err = config.RabbitMqConnect.PublishMessageWithRouteKey(config.AmqpGlobalConfig.Exchange["fanout"]["ticker"], "#", "text/plain", false, false, &t, amqp.Table{}, amqp.Transient, "")
		if err != nil {
			log.Println("{error:", err, "}")
		}
	}()
}

func buildOtherKLines(dataRedis redis.Conn, market *Market) {
	for _, period := range Lines {
		if period == 1 {
			continue
		}
		lock, _ := redis.String(dataRedis.Do("GET", (*market).UpdateLock(period)))
		if lock == "true" {
			return
		} else {
			dataRedis.Do("SETEX", (*market).UpdateLock(period), 10, "true")
		}
		payload := struct {
			MarketId int   `json:"market_id"`
			Period   int64 `json:"period"`
		}{MarketId: (*market).Id, Period: period}
		b, err := json.Marshal(payload)
		if err != nil {
			log.Println("error:", err)
		}
		err = config.RabbitMqConnect.PublishMessageWithRouteKey("quote.default", "quote.kLine.build", "text/plain", false, false, &b, amqp.Table{}, amqp.Transient, "")
		if err != nil {
			log.Println("{error:", err, "}")
		}
	}
}

func refreshTicker(dataRedis redis.Conn, market *Market) (ticker Ticker) {
	ticker.MarketId = (*market).Id
	now := time.Now()
	kJsons, _ := redis.Values(dataRedis.Do("ZRANGEBYSCORE", (*market).KLine(1), now.Add(-time.Hour*24).Unix(), now.Unix()))
	var k KLine
	for i, kJson := range kJsons {
		json.Unmarshal(kJson.([]byte), &k)
		if i == 0 {
			ticker.TickerAspect.Open = k.Open
		}
		ticker.TickerAspect.Close = k.Close
		if ticker.TickerAspect.Low.IsZero() || ticker.TickerAspect.Low.GreaterThan(k.Low) {
			ticker.TickerAspect.Low = k.Low
		}
		if ticker.TickerAspect.High.LessThan(k.High) {
			ticker.TickerAspect.High = k.High
		}
		ticker.TickerAspect.Vol = ticker.TickerAspect.Vol.Add(k.Vol)
	}
	return
}

func createSubQuote(quote *Quote) {
	if quote.QuoteCurrency.Source == "local" {
		return
	}
	b, err := json.Marshal(map[string]int{"id": quote.Id})
	if err != nil {
		log.Println(err)
	}
	err = initializers.PublishMessageWithRouteKey("quote.default", "quote.sub.build", "text/plain", &b, amqp.Table{}, amqp.Transient)
	if err != nil {
		log.Println(err)
	}
}
