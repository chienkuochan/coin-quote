package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	envConfig "quote/config"
	"quote/initializers"
	"quote/models"
	"quote/sourceData/binance"
	"quote/sourceData/huobi"
	"quote/utils"
)

func main() {
	initialize()

	go func() {
		err := huobi.GetHuobiPrice()
		for err != nil {
			err = huobi.GetHuobiPrice()
		}
	}()

	go func() {
		err := binance.GetBinancePrices()
		for err != nil {
			err = binance.GetBinancePrices()
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	time.Sleep(500)
	closeResource()
}

func initialize() {
	envConfig.InitEnv()
	utils.InitDB()
	utils.InitRedisPools()
	initializers.InitializeAmqpConfig()
	models.AutoMigrations()
	initializers.InitCacheData()

	setLog()
	err := os.MkdirAll("pids", 0755)
	if err != nil {
		log.Fatalf("create folder error: %v", err)
	}
	err = ioutil.WriteFile("pids/getData.pid", []byte(strconv.Itoa(os.Getpid())), 0644)
	if err != nil {
		log.Fatalf("open file error: %v", err)
	}

}

func closeResource() {
	initializers.CloseAmqpConnection()
	utils.CloseRedisPools()
	utils.CloseDB()
}

func setLog() {
	err := os.Mkdir("logs", 0755)
	if err != nil {
		if !os.IsExist(err) {
			log.Fatalf("create folder error: %v", err)
		}
	}
	file, err := os.OpenFile("logs/getData.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("open file error: %v", err)
	}
	log.SetOutput(file)
}
