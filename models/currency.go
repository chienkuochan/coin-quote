package models

import (
	"fmt"

	"github.com/jinzhu/gorm"

	"quote/utils"
)

type Currency struct {
	CommonModel
	Key     string  `json:"name"`    // 币的唯一标示
	Symbol  string  `json:"symbol"`  // 币的简称
	Source  string  `json:"source"`  // 币的来源，交易所
	Visible bool    `json:"visible"` // 是否可用
	Quotes  []Quote `json:"quotes" sql:"-"`
}

func InitAllCurrencies(db *utils.GormDB) {
	db.Where("visible = ?", true).Find(&AllCurrencies)
}

func FindCurrencyById(id int) (Currency, error) {
	for _, currency := range AllCurrencies {
		if currency.Id == id {
			return currency, nil
		}
	}
	var currency Currency
	return currency, fmt.Errorf("No currency can be found.")
}

func FindCurrencyBySymbol(symbol string) (Currency, error) {
	for _, currency := range AllCurrencies {
		if currency.Symbol == symbol {
			return currency, nil
		}
	}
	var currency Currency
	return currency, fmt.Errorf("No currency can be found.")
}

func (c *Currency) AfterFind(db *gorm.DB) {
}
