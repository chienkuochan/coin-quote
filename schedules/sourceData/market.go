package sourceData

import (
	. "quote/models"
	"quote/utils"
)

func CleanMarketExpiredKLine() {
	dataRedis := utils.GetRedisConn("data")
	defer dataRedis.Close()

	for _, market := range AllMarkets {
		for _, period := range Lines {
			dataRedis.Do("ZREMRANGEBYRANK", market.KLine(period), 0, -1000)
		}
	}
}

func CleanMarketExpiredPrice() {
	dataRedis := utils.GetRedisConn("data")
	defer dataRedis.Close()

	for _, market := range AllMarkets {
		dataRedis.Do("ZREMRANGEBYRANK", market.TimeLine(), 0, -1000)
	}
}
