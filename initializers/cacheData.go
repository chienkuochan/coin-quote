package initializers

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"quote/config"
	. "quote/models"
	"quote/utils"
)

var QueueName string

type Payload struct {
	Update string `json:"update"`
}

func InitCacheData() {
	db := utils.DbBegin()
	defer db.DbRollback()
	InitAllCurrencies(db)
	InitAllMarkets(db)
}

func LoadCacheData() {
	InitCacheData()
	go func() {
		channel, err := config.RabbitMqConnect.Channel()
		if err != nil {
			fmt.Errorf("Channel: %s", err)
		}
		channel.ExchangeDeclare(config.AmqpGlobalConfig.Exchange["fanout"]["default"], "fanout", true, false, false, false, nil)
		queue, err := channel.QueueDeclare("", true, true, false, false, nil)
		if err != nil {
			return
		}
		QueueName = queue.Name
		channel.QueueBind(queue.Name, QueueName, config.AmqpGlobalConfig.Exchange["fanout"]["default"], false, nil)
		msgs, _ := channel.Consume(queue.Name, "", true, false, false, false, nil)
		for d := range msgs {
			var payload Payload
			err := json.Unmarshal(d.Body, &payload)
			if err == nil {
				reflect.ValueOf(&payload).MethodByName(payload.Update).Call([]reflect.Value{})
			} else {
				log.Println(fmt.Sprintf("{error: %v}", err))
			}
		}
		return
	}()
}

func (payload *Payload) Currencies() {
	db := utils.DbBegin()
	defer db.DbRollback()
	InitAllCurrencies(db)
}

func (payload *Payload) Markets() {
	db := utils.DbBegin()
	defer db.DbRollback()
	InitAllMarkets(db)
}
