package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"time"

	sneaker "github.com/oldfritter/sneaker-go/v3"

	envConfig "quote/config"
	"quote/initializers"
	"quote/utils"
)

func main() {
	initialize()
	initializers.InitWorkers()

	StartAllWorkers()

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
	initializers.InitCacheData()

	setLog()
	err := os.MkdirAll("pids", 0755)
	if err != nil {
		log.Fatalf("create folder error: %v", err)
	}
	err = ioutil.WriteFile("pids/workers.pid", []byte(strconv.Itoa(os.Getpid())), 0644)
	if err != nil {
		log.Fatalf("open file error: %v", err)
	}
}

func closeResource() {
	initializers.CloseAmqpConnection()
	utils.CloseDB()
}

func StartAllWorkers() {
	for _, w := range envConfig.AllWorkerIs {
		for i := 0; i < w.GetThreads(); i++ {
			go func(w sneaker.WorkerI) {
				w.InitLogger()
				sneaker.SubscribeMessageByQueue(envConfig.RabbitMqConnect.Connection, w, nil)
			}(w)
		}
	}
}

func setLog() {
	err := os.Mkdir("logs", 0755)
	if err != nil {
		if !os.IsExist(err) {
			log.Fatalf("create folder error: %v", err)
		}
	}
	file, err := os.OpenFile("logs/workers.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("open file error: %v", err)
	}
	log.SetOutput(file)
}
