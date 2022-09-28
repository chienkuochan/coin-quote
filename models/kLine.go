package models

import (
	"github.com/shopspring/decimal"
)

type KLine struct {
	Timestamp int64           `json:"timestamp"` // 时间戳
	Open      decimal.Decimal `json:"open"`      // 开盘价
	Close     decimal.Decimal `json:"close"`     // 收盘价
	Low       decimal.Decimal `json:"low"`       // 最低价
	High      decimal.Decimal `json:"high"`      // 最高价
	Vol       decimal.Decimal `json:"vol"`       // 量
}

type NotifyKLine struct {
	MarketId int   `json:"market_id"` // 市场
	Period   int64 `json:"period"`    // 所属K线，1分钟，5分钟，或者其他
	KLine    KLine `json:"k_line"`    // K线数据
}
