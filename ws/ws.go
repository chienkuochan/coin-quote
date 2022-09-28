package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	envConfig "quote/config"
	"quote/models"
	"quote/utils"
	"quote/ws/v1"
)

func main() {
	initialize()

	http.HandleFunc("/ws", v1.Quotes)
	log.Fatal(http.ListenAndServe("127.0.0.1:9010", nil))

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
	models.AutoMigrations()

	setLog()
	err := os.MkdirAll("pids", 0755)
	if err != nil {
		log.Fatalf("create folder error: %v", err)
	}
	err = ioutil.WriteFile("pids/ws.pid", []byte(strconv.Itoa(os.Getpid())), 0644)
	if err != nil {
		log.Fatalf("open file error: %v", err)
	}

}

func closeResource() {
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
	file, err := os.OpenFile("logs/ws.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("open file error: %v", err)
	}
	log.SetOutput(file)
}
