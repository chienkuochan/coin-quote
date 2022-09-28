package baseRate

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/shopspring/decimal"

	. "quote/models"
	"quote/utils"
)

func UsdtToCny() {
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
	var cny Currency
	db.FirstOrInit(&cny, map[string]interface{}{
		"key":     "CNY",
		"symbol":  "cny",
		"source":  "local",
		"visible": true,
	})
	db.Save(&cny)
	var usdts []Currency
	db.Where("symbol = ?", "usdt").Find(&usdts)
	for _, u := range usdts {
		var quote Quote
		db.FirstOrInit(&quote, map[string]interface{}{
			"base_id":  u.Id,
			"quote_id": cny.Id,
			"type":     "Quotes::" + strings.Title(u.Source),
		})
		quote.Timestamp = time.Now().UnixNano() / 1000000
		quote.Price = price
		db.Save(&quote)
	}
	db.DbCommit()
}

func HuobiUsdtToCny() {
	url := "https://otc-api-hk.eiijo.cn/v1/data/trade-market?coinId=2&currency=1&tradeType=sell&currPage=1&payMethod=0&country=37&blockType=general&online=1&range=0&amount="
	ctx, cancelFun := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFun()
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		log.Println(err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	var result struct {
		Data []struct {
			Price decimal.Decimal `json:"price"`
		} `json:"data"`
	}
	json.Unmarshal(body, &result)
	db := utils.DbBegin()
	defer db.DbRollback()
	var cny Currency
	db.FirstOrInit(&cny, map[string]interface{}{
		"key":     "CNY",
		"symbol":  "cny",
		"source":  "local",
		"visible": true,
	})
	db.Save(&cny)
	var usdts []Currency
	db.Where("symbol = ?", "usdt").Where("source in (?)", []string{"huobi", "watbtc"}).Find(&usdts)
	for _, usdt := range usdts {
		var quote Quote
		db.FirstOrInit(&quote, map[string]interface{}{
			"base_id":  usdt.Id,
			"quote_id": cny.Id,
			"type":     "Quotes::" + strings.Title(usdt.Source),
		})
		quote.Timestamp = time.Now().UnixNano() / 1000000
		quote.Price = result.Data[0].Price
		quote.Source = usdt.Source
		db.Save(&quote)
	}
	db.DbCommit()
}

func BinanceUsdtToCny() {
	url := "https://c2c.binance.com/gateway-api/v2/public/c2c/adv/search"
	ctx, cancelFun := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFun()
	req, _ := http.NewRequest(http.MethodPost, url, strings.NewReader("{\"rows\":1,\"fiat\":\"CNY\",\"page\":1,\"asset\":\"USDT\",\"tradeType\":\"BUY\"}"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		log.Println(err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	var result struct {
		Data []struct {
			AdvDetail struct {
				Price decimal.Decimal `json:"price"`
			} `json:"advDetail"`
		} `json:"data"`
	}
	json.Unmarshal(body, &result)
	db := utils.DbBegin()
	defer db.DbRollback()
	var cny Currency
	db.FirstOrInit(&cny, map[string]interface{}{
		"key":     "CNY",
		"symbol":  "cny",
		"source":  "local",
		"visible": true,
	})
	db.Save(&cny)
	var usdt Currency
	db.Where("symbol = ?", "usdt").Where("source = ?", "binance").First(&usdt)
	var quote Quote
	db.FirstOrInit(&quote, map[string]interface{}{
		"base_id":  usdt.Id,
		"quote_id": cny.Id,
		"type":     "Quotes::" + strings.Title(usdt.Source),
	})
	quote.Timestamp = time.Now().UnixNano() / 1000000
	quote.Price = result.Data[0].AdvDetail.Price
	quote.Source = usdt.Source
	db.Save(&quote)
	db.DbCommit()
}
