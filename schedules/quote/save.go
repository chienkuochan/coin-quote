package quote

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gomodule/redigo/redis"

	. "quote/models"
	"quote/utils"
)

func SaveDataFromRedis() {
	db := utils.Db
	var markets []Market
	db.Where("visible = ?", true).Find(&markets)

	dataRedis := utils.GetRedisConn("data")
	defer dataRedis.Close()

	for _, market := range markets {
		keys, err := redis.Strings(dataRedis.Do("KEYS", fmt.Sprintf("Quotes:%d:*", market.Id)))
		if err != nil {
		}
		for _, key := range keys {

			qByte, err := redis.Bytes(dataRedis.Do("GET", key))
			if err != nil {
			}
			var market Market
			var simple SimpleQuote
			var baseC, quoteC Currency
			var quoteCs []Currency
			json.Unmarshal(qByte, &simple)

			m := utils.DbBegin()
			defer m.DbRollback()

			if m.Where("source = ?", simple.Source).Where("symbol = ?", simple.Market).First(&market).RecordNotFound() {
				market.Visible = true
				market.Source = simple.Source
				market.Symbol = simple.Market
				market.Name = strings.ToUpper(simple.Market)
			}
			if m.Where("source = ?", simple.Source).Where("symbol = ?", simple.Base).First(&baseC).RecordNotFound() {
				baseC.Visible = true
				baseC.Source = simple.Source
				baseC.Symbol = simple.Base
				baseC.Key = strings.ToUpper(simple.Base)
				db.Save(&baseC)
			}
			if m.Where("source in (?)", []string{simple.Source, "local"}).Where("symbol = ?", simple.Quote).Find(&quoteCs).RecordNotFound() {
				quoteC.Visible = true
				quoteC.Source = simple.Source
				quoteC.Symbol = simple.Quote
				quoteC.Key = strings.ToUpper(simple.Quote)
				db.Save(&quoteC)
				market.QuoteId = quoteC.Id
			}
			if len(quoteCs) > 0 {
				market.QuoteId = quoteCs[0].Id
				for _, quoteC := range quoteCs {
					var quote Quote
					if m.Where("source = ?", simple.Source).Where("base_id = ?", baseC.Id).Where("quote_id = ?", quoteC.Id).Where("market_id = ?", market.Id).First(&quote).RecordNotFound() {
						quote.Type = "Quotes::" + strings.Title(simple.Source)
						quote.Source = simple.Source
						quote.BaseId = baseC.Id
						quote.QuoteId = quoteC.Id
						quote.MarketId = market.Id
					}
					quote.Price = simple.Price
					quote.Timestamp = simple.Timestamp
					m.Save(&quote)
				}
			}
			if market.Id == 0 {
				market.BaseId = baseC.Id
				db.Save(&market)
			}
			m.DbCommit()
		}
	}
}
