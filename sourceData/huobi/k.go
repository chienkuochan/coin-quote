package huobi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"

	"quote/models"
	"quote/sourceData/redis"
	sourceDataUtils "quote/sourceData/utils"
	"quote/utils"
)

type HuobiPayload struct {
	Ch   string `json:"ch"`
	Ts   int64  `json:"ts"`
	Tick struct {
		Id     int             `json:"id"`
		Open   decimal.Decimal `json:"open"`
		Close  decimal.Decimal `json:"close"`
		Low    decimal.Decimal `json:"low"`
		High   decimal.Decimal `json:"high"`
		Amount decimal.Decimal `json:"amount"`
		Vol    decimal.Decimal `json:"vol"`
		Count  int             `json:"count"`
	} `json:"tick"`
}

func GetHuobiPrice() error {
	u := url.URL{Scheme: "wss", Host: "api.huobi.pro", Path: "/ws"}
	log.Println("connecting to ", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Println("dial:", err)
		return err
	}
	huobiMarkets := models.FindMarketsBySource("huobi")
	for _, m := range huobiMarkets {
		err = c.WriteMessage(websocket.TextMessage, []byte("{\"sub\":\"market."+m.Symbol+".kline.1min\",\"id\":\"oldfritter\"}"))
		if err != nil {
			log.Println("write:", err)
			return err
		}
	}
	defer c.Close()
	errChan := make(chan error)
	reloadChan := make(chan error)
	go func(chan error) {
		sourceDataUtils.ListenReloadMarkets(reloadChan)
	}(reloadChan)
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				errChan <- err
				return
			}
			strByte, _ := utils.GzipDecode(message)
			var payload HuobiPayload
			json.Unmarshal(strByte, &payload)
			if payload.Ch != "" {
				kLine := models.KLine{
					Timestamp: payload.Ts,
					Open:      payload.Tick.Open,
					Close:     payload.Tick.Close,
					Low:       payload.Tick.Low,
					High:      payload.Tick.High,
					Vol:       payload.Tick.Vol,
				}
				symbol := strings.Replace(payload.Ch, "market.", "", 1)
				symbol = strings.Replace(symbol, ".kline.1min", "", 1)
				for _, m := range huobiMarkets {
					if m.Symbol == symbol {
						redis.Save(&m, &kLine)
					}
				}
			} else {
				var ping struct {
					Ping int64 `json:"ping"`
				}
				b, err := json.Marshal(&ping)
				if err != nil {
					log.Println("write:", err)
					continue
				}
				err = c.WriteMessage(websocket.TextMessage, b)
				if err != nil {
					log.Println("write:", err)
					continue
				}
			}
		}
	}()

	for {
		select {
		case <-errChan:
			return fmt.Errorf("connect is closed!")
		case <-reloadChan:
			return fmt.Errorf("connect need reopen!")
		}
	}
	return nil
}
