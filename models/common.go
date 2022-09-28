package models

import (
	"time"

	"quote/utils"
)

var (
	AllCurrencies []Currency
	AllMarkets    []Market

	Lines = []int64{1, 5, 15, 30, 60}
)

type CommonModel struct {
	Id        int       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func AutoMigrations() {
	mainDB := utils.DbBegin()
	defer mainDB.DbRollback()

	// currency
	mainDB.AutoMigrate(&Currency{})
	mainDB.Model(&Currency{}).AddIndex("index_currencies_on_visible", "visible")

	// market
	mainDB.AutoMigrate(&Market{})
	mainDB.Model(&Market{}).AddUniqueIndex("index_markets_on_name", "name", "source")

	// quote
	mainDB.AutoMigrate(&Quote{})
	mainDB.Model(&Quote{}).AddUniqueIndex("index_quotes_on_base_id_market_id_and_and_quote_id", "base_id", "market_id", "quote_id")
}
