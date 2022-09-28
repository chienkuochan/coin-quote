package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"quote/routes"
	"quote/utils"
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	newrelic "github.com/oldfritter/echo-middleware"

	"quote/initializers"
)

func main() {
	initialize()

	e := echo.New()

	e.Use(newrelic.NewRelic("Quote", "f695fe25ce2fe9fe93fc4003b2311df889507ca9"))

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(initializers.Filter)

	routes.SetQuoteInterfaces(e)

	e.HTTPErrorHandler = customHTTPErrorHandler

	// Start server
	e.HideBanner = true
	// Start server
	go func() {
		if err := e.Start(":9091"); err != nil {
			closeResource()
			e.Logger.Info("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 10 seconds.
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

func customHTTPErrorHandler(err error, context echo.Context) {
	if _, ok := err.(utils.Response); ok {
		context.JSON(http.StatusBadRequest, err)
	} else {
		// panic(err)
	}
}

func initialize() {
	utils.InitDB()
	utils.InitRedisPools()
	initializers.InitCacheData()
	setLog()

	// 记录 pid
	err := ioutil.WriteFile("pids/api.pid", []byte(strconv.Itoa(os.Getpid())), 0644)
	if err != nil {
		log.Println(err)
	}
}

func closeResource() {
	utils.CloseDB()
}

func setLog() {
	err := os.Mkdir("logs", 0755)
	if err != nil {
		if !os.IsExist(err) {
			log.Fatalf("create folder error: %v", err)
		}
	}
	file, err := os.OpenFile("logs/api.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("open file error: %v", err)
	}
	log.SetOutput(file)
}
