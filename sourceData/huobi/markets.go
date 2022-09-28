package huobi

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	. "quote/models"
	"quote/utils"
)

type HuobiResponse struct {
	Data []struct {
		Symbol string `json:"symbol"`
		Status bool   `json:"visit-enabled"`
		Base   string `json:"base-currency"`
		Quote  string `json:"quote-currency"`
	} `json:"data"`
}

func GetMarkets() {
	ctx, cancelFun := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFun()
	req, err := http.NewRequest(http.MethodGet, "https://api.huobi.com/v1/settings/symbols?r=o58hsjzm4m", nil)
	if err != nil {
		log.Println(err)
		return
	}
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		log.Println(err)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	var huobiResponse HuobiResponse
	err = json.Unmarshal(body, &huobiResponse)
	if err != nil {
		log.Println(err)
		return
	}
	mdb := utils.DbBegin()
	defer mdb.DbRollback()
	var markets []Market
	mdb.Model(Market{}).Where("source = ?", "huobi").Updates(Market{Visible: false})
	mdb.DbRollback()
	for _, symbol := range huobiResponse.Data {
		db := utils.DbBegin()
		defer db.DbRollback()
		if symbol.Status {
			db.Where("source = ?", "huobi").Find(&markets)
			var market Market
			db.FirstOrInit(&market, map[string]interface{}{"name": strings.ToUpper(symbol.Symbol), "symbol": symbol.Symbol, "source": "huobi"})
			var base, quote Currency
			db.FirstOrInit(&base, map[string]interface{}{"symbol": symbol.Base, "key": strings.ToUpper(symbol.Base), "visible": true, "source": "huobi"})
			db.FirstOrInit(&quote, map[string]interface{}{"symbol": symbol.Quote, "key": strings.ToUpper(symbol.Quote), "visible": true, "source": "huobi"})
			db.Save(&base)
			db.Save(&quote)
			market.BaseId = base.Id
			market.QuoteId = quote.Id
			market.Visible = true
			db.Save(&market)
		}
		db.DbCommit()
	}
	return
}
