package v1

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	. "quote/models"
	"quote/utils"
)

func Quotes(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{}
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("upgrade:", err)
		return
	}
	c.SetWriteDeadline(time.Now().Add(time.Hour * 24))
	defer c.Close()
	var params struct {
		Symbols    []string `json:"symbols"`
		Sources    []string `json:"sources"`
		Currencies []string `json:"currencies"`
	}
	_, m, err := c.ReadMessage()
	json.Unmarshal(m, &params)
	if len(params.Symbols) == 0 {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	err = utils.ListenPubSubChannels(
		ctx,
		func() error {
			return nil
		},
		func(channel string, m []byte) error {
			var message struct {
				Timestamp int64  `json:"timestamp"`
				Source    string `json:"source"`
				Market    string `json:"market"`
				Base      string `json:"base"`
				Quote     string `json:"quote"`
				Price     string `json:"price"`
			}
			if channel == RedisNotify {
				json.Unmarshal(m, &message)
				var inBase, inSource, inQuote bool
				for _, s := range params.Symbols {
					if s == message.Base {
						inBase = true
					}
				}
				if !inBase {
					return nil
				}
				if len(params.Sources) == 0 {
					inSource = true
				} else {
					for _, s := range params.Sources {
						if s == message.Source {
							inSource = true
						}
					}
				}
				if !inSource {
					return nil
				}
				if len(params.Currencies) == 0 {
					inQuote = true
				} else {
					for _, s := range params.Currencies {
						if s == message.Quote {
							inQuote = true
						}
					}
				}
				if !inQuote {
					return nil
				}
				err := c.WriteMessage(websocket.TextMessage, m)
				if err != nil {
					log.Println("write:", err)
					cancel()
				}
			}
			return nil
		},
		RedisNotify,
	)
	if err != nil {
		log.Println(err)
		return
	}
}
