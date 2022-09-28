package baseRate

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	. "quote/models"
	"quote/utils"
)

func CnyToUsd() {
	url := "https://markets.money.cnn.com/common/modules/iframe/currencyConverter.asp?convert=1&amount=1&base=CNY&quote=USD"
	ctx, cancelFun := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFun()
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		log.Println(err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	p := strings.TrimLeft(string(body), "= ")
	p = strings.TrimRight(p, " USD")
	price, _ := decimal.NewFromString(p)
	db := utils.DbBegin()
	defer db.DbRollback()
	var usd, cny, cnst Currency
	db.FirstOrInit(&usd, map[string]interface{}{
		"key":     "USD",
		"symbol":  "usd",
		"source":  "local",
		"visible": true,
	})
	db.Save(&usd)
	db.FirstOrInit(&cny, map[string]interface{}{
		"key":     "CNY",
		"symbol":  "cny",
		"source":  "local",
		"visible": true,
	})
	db.Save(&cny)
	db.FirstOrInit(&cnst, map[string]interface{}{
		"key":     "CNST",
		"symbol":  "cnst",
		"source":  "local",
		"visible": true,
	})
	db.Save(&cnst)
	var quote, cnstQuote Quote
	db.FirstOrInit(&quote, map[string]interface{}{
		"base_id":  cny.Id,
		"quote_id": usd.Id,
		"source":   "local",
		"type":     "Quotes::Local",
	})
	quote.Timestamp = time.Now().UnixNano() / 1000000
	quote.Price = price
	db.Save(&quote)
	db.FirstOrInit(&cnstQuote, map[string]interface{}{
		"base_id":  cnst.Id,
		"quote_id": usd.Id,
		"source":   "local",
		"type":     "Quotes::Local",
	})
	cnstQuote.Timestamp = time.Now().UnixNano() / 1000000
	cnstQuote.Price = price
	db.Save(&cnstQuote)
	db.DbCommit()
}

func UsdToCny() {
	url := "https://markets.money.cnn.com/common/modules/iframe/currencyConverter.asp?convert=1&amount=1&base=USD&quote=CNY"
	ctx, cancelFun := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFun()
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		log.Println(err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	p := strings.TrimLeft(string(body), "= ")
	p = strings.TrimRight(p, " CNY")
	price, _ := decimal.NewFromString(p)
	db := utils.DbBegin()
	defer db.DbRollback()
	var usd, cny, cnst Currency
	db.FirstOrInit(&usd, map[string]interface{}{
		"key":     "USD",
		"symbol":  "usd",
		"source":  "local",
		"visible": true,
	})
	db.Save(&usd)
	db.FirstOrInit(&cny, map[string]interface{}{
		"key":     "CNY",
		"symbol":  "cny",
		"source":  "local",
		"visible": true,
	})
	db.Save(&cny)
	db.FirstOrInit(&cnst, map[string]interface{}{
		"key":     "CNST",
		"symbol":  "cnst",
		"source":  "local",
		"visible": true,
	})
	db.Save(&cnst)
	var quote, cnstQuote Quote
	db.FirstOrInit(&quote, map[string]interface{}{
		"base_id":  usd.Id,
		"quote_id": cny.Id,
		"source":   "local",
		"type":     "Quotes::Local",
	})
	quote.Timestamp = time.Now().UnixNano() / 1000000
	quote.Price = price
	db.Save(&quote)
	db.FirstOrInit(&cnstQuote, map[string]interface{}{
		"base_id":  usd.Id,
		"quote_id": cnst.Id,
		"source":   "local",
		"type":     "Quotes::Local",
	})
	cnstQuote.Timestamp = time.Now().UnixNano() / 1000000
	cnstQuote.Price = price
	db.Save(&cnstQuote)
	db.DbCommit()
}
