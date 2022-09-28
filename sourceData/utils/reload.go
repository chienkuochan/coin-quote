package utils

import (
	"encoding/json"
	"log"

	"quote/config"
	"quote/initializers"
)

func ListenReloadMarkets(reloadChan chan error) {
	channel, err := config.RabbitMqConnect.Channel()
	if err != nil {
		log.Println(err)
	}
	channel.ExchangeDeclare(config.AmqpGlobalConfig.Exchange["fanout"]["default"], "fanout", true, false, false, false, nil)
	queue, err := channel.QueueDeclare("", true, true, false, false, nil)
	if err != nil {
		log.Println(err)
		return
	}
	channel.QueueBind(queue.Name, queue.Name, config.AmqpGlobalConfig.Exchange["fanout"]["default"], false, nil)
	msgs, _ := channel.Consume(queue.Name, "", true, false, false, false, nil)
	for d := range msgs {
		var payload initializers.Payload
		err := json.Unmarshal(d.Body, &payload)
		if err == nil && payload.Update == "Markets" {
			reloadChan <- err
		}
	}
}
