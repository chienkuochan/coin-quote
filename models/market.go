package models

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"quote/utils"
)

type Market struct {
	CommonModel
	BaseId  int          `json:"base_id"`                         // 商品币种
	QuoteId int          `json:"quote_id"`                        // 计价币种
	Name    string       `json:"name" gorm:"type:varchar(64)"`    // 市场名称
	Source  string       `json:"source" gorm:"type:varchar(16);"` // 源，huobi或者binance
	Symbol  string       `json:"symbol" gorm:"type:varchar(16)"`  // 源市场的唯一标示
	Visible bool         `json:"visible"`                         // 是否可用
	Ticker  TickerAspect `json:"-" sql:"-"`                       // 最新行情
}

func InitAllMarkets(db *utils.GormDB) {
	db.Where("visible = ?", true).Find(&AllMarkets)
}

func FindMarketById(id int) (market Market, err error) {
	for i, _ := range AllMarkets {
		if AllMarkets[i].Id == id {
			market = AllMarkets[i]
			return
		}
	}
	err = fmt.Errorf("No market can be found.")
	return
}

func FindMarketByName(name string) (market Market, err error) {
	for _, m := range AllMarkets {
		if m.Name == name {
			market = m
			return
		}
	}
	err = fmt.Errorf("No market can be found.")
	return
}

func FindMarketsBySource(source string) (markets []Market) {
	for _, market := range AllMarkets {
		if market.Source == source {
			markets = append(markets, market)
		}
	}
	return
}

func (market *Market) AfterFind(db *gorm.DB) {
}

func (market *Market) TimeLine() string {
	return fmt.Sprintf("market:timeLine:%v", market.Id)
}

func (market *Market) KLine(period int64) string {
	return fmt.Sprintf("market:kLine:%v:%v", market.Id, period)
}

func (market *Market) KLineNotify(period int64) string {
	return "market:kLine:notify"
}

func (market *Market) TickerNotify() string {
	return fmt.Sprintf("market:ticker:notify:%v", market.Id)
}

func (market *Market) UpdateLock(period int64) string {
	return fmt.Sprintf("market:update:lock:%v:%v", market.Id, period)
}
