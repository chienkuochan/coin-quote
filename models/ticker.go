package models

import (
	"github.com/shopspring/decimal"
)

type TickerAspect struct {
	Open  decimal.Decimal `json:"open"`  // 24h开盘价
	Close decimal.Decimal `json:"close"` // 24h收盘价
	Low   decimal.Decimal `json:"low"`   // 24h最低价
	High  decimal.Decimal `json:"high"`  // 24h最高价
	Vol   decimal.Decimal `json:"vol"`   // 24h交易量
}
type Ticker struct {
	MarketId     int          `json:"market_id"`
	TickerAspect TickerAspect `json:"ticker"`
}
