package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/robfig/cron/v3"

	envConfig "quote/config"
	"quote/initializers"
	"quote/models"
	"quote/schedules/baseRate"
	"quote/schedules/quote"
	"quote/schedules/rfinex"
	"quote/schedules/sourceData"
	"quote/schedules/watbtc"
	"quote/sourceData/binance"
	"quote/sourceData/huobi"
	"quote/utils"
)

func main() {
	initialize()

	setLog()
	err := ioutil.WriteFile("pids/schedule.pid", []byte(strconv.Itoa(os.Getpid())), 0644)
	if err != nil {
		fmt.Println(err)
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	closeResource()
}

func initialize() {
	envConfig.InitEnv()
	utils.InitDB()
	utils.InitAwsS3Config()
	utils.InitRedisPools()
	initializers.InitializeAmqpConfig()
	models.AutoMigrations()
	initializers.LoadCacheData()
	InitSchedule()

	initializers.InitWorkers()
}

func closeResource() {
	initializers.CloseAmqpConnection()
	utils.CloseRedisPools()
	utils.CloseDB()
}

func InitSchedule() {
	c := cron.New(cron.WithSeconds())

	// 清理过期价格
	c.AddFunc("0 59 23 * * *", sourceData.CleanMarketExpiredKLine)
	c.AddFunc("0 59 23 * * *", sourceData.CleanMarketExpiredPrice)

	c.AddFunc("0 * * * * *", binance.GetMarkets)
	c.AddFunc("0 * * * * *", huobi.GetMarkets)

	c.AddFunc("0 * * * * *", baseRate.CnyToUsd)
	c.AddFunc("0 * * * * *", baseRate.UsdToCny)
	c.AddFunc("0 * * * * *", baseRate.UsdtToUsd)
	c.AddFunc("0 * * * * *", baseRate.HuobiUsdtToCny)
	c.AddFunc("0 * * * * *", baseRate.BinanceUsdtToCny)

	c.AddFunc("*/30 * * * * *", rfinex.GetRfinexTickers)
	c.AddFunc("*/30 * * * * *", watbtc.GetWatbtcTickers)

	c.AddFunc("*/30 * * * * *", quote.SaveDataFromRedis)

	c.Start()
}

func setLog() {
	err := os.Mkdir("logs", 0755)
	if err != nil {
		if !os.IsExist(err) {
			log.Fatalf("create folder error: %v", err)
		}
	}
	file, err := os.OpenFile("logs/schedule.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("open file error: %v", err)
	}
	log.SetOutput(file)
}
