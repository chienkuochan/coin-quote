package initializers

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	"github.com/oldfritter/sneaker-go/v3"
	"github.com/streadway/amqp"
	"gopkg.in/yaml.v2"

	"quote/config"
)

func InitializeAmqpConfig() {
	path_str, _ := filepath.Abs("config/amqp.yml")
	content, err := ioutil.ReadFile(path_str)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = yaml.Unmarshal(content, &config.AmqpGlobalConfig)
	if err != nil {
		log.Fatal(err)
		return
	}
	InitializeAmqpConnection()
}

func InitializeAmqpConnection() {
	var err error
	conn, err := amqp.Dial("amqp://" + config.AmqpGlobalConfig.Connect.Username + ":" + config.AmqpGlobalConfig.Connect.Password + "@" + config.AmqpGlobalConfig.Connect.Host + ":" + config.AmqpGlobalConfig.Connect.Port + "/" + config.AmqpGlobalConfig.Connect.Vhost)
	config.RabbitMqConnect = sneaker.RabbitMqConnect{conn}
	if err != nil {
		time.Sleep(5000)
		InitializeAmqpConnection()
		return
	}
	go func() {
		<-config.RabbitMqConnect.NotifyClose(make(chan *amqp.Error))
		InitializeAmqpConnection()
	}()
}

func CloseAmqpConnection() {
	config.RabbitMqConnect.Close()
}

func GetRabbitMqConnect() sneaker.RabbitMqConnect {
	return config.RabbitMqConnect
}

func PublishMessageWithRouteKey(exchange, routeKey, contentType string, message *[]byte, arguments amqp.Table, deliveryMode uint8) error {
	channel, err := config.RabbitMqConnect.Channel()
	defer channel.Close()
	if err != nil {
		return fmt.Errorf("Channel: %s", err)
	}
	if err = channel.Publish(
		exchange, // publish to an exchange
		routeKey, // routing to 0 or more queues
		false,    // mandatory
		false,    // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     contentType,
			ContentEncoding: "",
			Body:            *message,
			DeliveryMode:    deliveryMode, // amqp.Persistent, amqp.Transient // 1=non-persistent, 2=persistent
			Priority:        0,            // 0-9
			// a bunch of application/implementation-specific fields
		},
	); err != nil {
		return fmt.Errorf("Queue Publish: %s", err)
	}
	return nil
}

func PublishMessageToQueue(queue, contentType string, message *[]byte, arguments amqp.Table, deliveryMode uint8) error {
	channel, err := config.RabbitMqConnect.Channel()
	defer channel.Close()
	if err != nil {
		return fmt.Errorf("Channel: %s", err)
	}
	if err = channel.Publish(
		"",    // publish to an exchange
		queue, // routing to 0 or more queues
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     contentType,
			ContentEncoding: "",
			Body:            *message,
			DeliveryMode:    deliveryMode, // amqp.Persistent, amqp.Transient // 1=non-persistent, 2=persistent
			Priority:        0,            // 0-9
			// a bunch of application/implementation-specific fields
		},
	); err != nil {
		return fmt.Errorf("Queue Publish: %s", err)
	}
	return nil
}
