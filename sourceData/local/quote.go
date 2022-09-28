package local

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

func GetLocalQuote() error {
	u := url.URL{Scheme: "wss", Host: "quote.bitip.in", Path: "/ws"}
	log.Println("connecting to ", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Println("dial:", err)
		return err
	}
	err = c.WriteMessage(websocket.TextMessage, []byte("{\"symbols\":[\"rmc\"],\"sources\":[\"watbtc\"],\"currencies\":[],\"id\":\"oldfritter\"}"))
	if err != nil {
		log.Println("write:", err)
		return err
	}
	defer c.Close()
	errChan := make(chan error)
	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				errChan <- err
				return
			}
			log.Println("payload:", string(message))
		}
	}()

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-errChan:
			return fmt.Errorf("connect is closed!")
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return err
			}
		}
	}
	return nil
}
