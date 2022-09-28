package baseRate

// import (
//   "strings"
//   "time"
//
//   "github.com/shopspring/decimal"
//
//   . "quote/models"
//   "quote/utils"
// )

func UsdtToUsd() {
	// price := decimal.NewFromFloat(1)
	// db := utils.DbBegin()
	// defer db.DbRollback()
	// var usd Currency
	// db.FirstOrInit(&usd, map[string]interface{}{
	//   "key":     "USD",
	//   "symbol":  "usd",
	//   "source":  "local",
	//   "visible": true,
	// })
	// db.Save(&usd)
	// var usdts []Currency
	// db.Where("symbol = ?", "usdt").Find(&usdts)
	// for _, u := range usdts {
	//   var quote Quote
	//   db.FirstOrInit(&quote, map[string]interface{}{
	//     "base_id":  u.Id,
	//     "quote_id": usd.Id,
	//     "type":     "Quotes::" + strings.Title(u.Source),
	//   })
	//   quote.Timestamp = time.Now().UnixNano() / 1000000
	//   quote.Price = price
	//   db.Save(&quote)
	// }
	// db.DbCommit()
}
