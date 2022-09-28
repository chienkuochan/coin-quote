package sneakerWorkers

import (
	"encoding/json"
	"log"
	"time"

	// "github.com/gomodule/redigo/redis"
	sneaker "github.com/oldfritter/sneaker-go/v3"
	"github.com/streadway/amqp"

	"quote/config"
	. "quote/models"
	"quote/utils"
)

func InitializeSubQuoteBuildWorker() {
	for _, w := range config.AllWorkers {
		if w.Name == "SubQuoteBuildWorker" {
			config.AllWorkerIs = append(config.AllWorkerIs, &SubQuoteBuildWorker{w})
			return
		}
	}
}

type SubQuoteBuildWorker struct {
	sneaker.Worker
}

func (worker *SubQuoteBuildWorker) Work(payloadJson *[]byte) (err error) {
	start := time.Now().UnixNano()
	var payload struct {
		Id    int `json:"id"`
		Level int `json:"level"`
	}
	json.Unmarshal([]byte(*payloadJson), &payload)
	db := utils.DbBegin()
	defer db.DbRollback()
	var origin Quote
	if db.Where("id = ?", payload.Id).First(&origin).RecordNotFound() {
		return
	}
	var quotes []Quote
	if db.Where("`source` in (?)", []string{origin.Source, "local"}).Where("base_id = ?", origin.QuoteId).Where("market_id <> ?", origin.MarketId).Find(&quotes).RecordNotFound() {
		return
	}
	db.DbRollback()
	for _, q := range quotes {
		sub, err := worker.subQuote(&origin, &q)
		if err != nil {
			continue
		}
		if payload.Level < 3 && (sub.IsLegal() || sub.IsAnchored()) {
			worker.createSubQuote(&sub, payload.Level)
		}
	}
	worker.LogInfo(" payload: ", payload, ", time:", (time.Now().UnixNano()-start)/1000000, " ms")
	return
}

func (worker *SubQuoteBuildWorker) subQuote(origin, q *Quote) (Quote, error) {
	db := utils.DbBegin()
	defer db.DbRollback()
	var subQuote Quote
	subQuote.Type = origin.Type
	subQuote.BaseId = origin.BaseId
	subQuote.MarketId = origin.MarketId
	subQuote.Source = origin.Source
	subQuote.QuoteId = q.QuoteId
	subQuote.Price = origin.Price.Mul(q.Price)
	subQuote.Timestamp = origin.Timestamp
	subQuote.SetAttrs()
	db.Save(&subQuote)
	db.DbCommit()
	subQuote.SaveToRedis()
	subQuote.NotifyQuote()
	return subQuote, nil
}

func (worker *SubQuoteBuildWorker) createSubQuote(quote *Quote, level int) {
	b, err := json.Marshal(struct {
		Id    int `json:"id"`
		Level int `json:"level"`
	}{
		Id:    quote.Id,
		Level: level + 1,
	})
	if err != nil {
		log.Println(err)
	}
	err = config.RabbitMqConnect.PublishMessageWithRouteKey("quote.default", "quote.sub.build", "text/plain", false, false, &b, amqp.Table{}, amqp.Transient, "")
	if err != nil {
		log.Println(err)
	}
}
