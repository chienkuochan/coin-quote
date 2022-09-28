package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"

	envConfig "quote/config"
	"quote/utils"
)

func main() {

	envConfig.InitEnv()
	utils.InitRedisPools()

	ctx, cancel := context.WithCancel(context.Background())
	err := listenPubSubChannels(ctx,
		func() error {
			return nil
		},
		func(channel string, message []byte) error {
			fmt.Printf("channel: %s, message: %s\n", channel, message)

			// For the purpose of this example, cancel the listener's context
			// after receiving last message sent by publish().
			if string(message) == "goodbye" {
				cancel()
			}
			return nil
		},
		"c1")

	if err != nil {
		fmt.Println(err)
		return
	}
}

func listenPubSubChannels(ctx context.Context, onStart func() error, onMessage func(channel string, data []byte) error, channels ...string) (err error) {
	dataRedis := utils.GetRedisConn("data")
	defer dataRedis.Close()
	psc := redis.PubSubConn{Conn: dataRedis}
	if err := psc.Subscribe(redis.Args{}.AddFlat(channels)...); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() {
		for {
			switch n := psc.Receive().(type) {
			case error:
				done <- n
				return
			case redis.Message:
				if err := onMessage(n.Channel, n.Data); err != nil {
					done <- err
					return
				}
			case redis.Subscription:
				switch n.Count {
				case len(channels):
					if err := onStart(); err != nil {
						done <- err
						return
					}
				case 0:
					done <- nil
					return
				}
			}
		}
	}()

	const healthCheckPeriod = time.Minute
	ticker := time.NewTicker(healthCheckPeriod)
	defer ticker.Stop()
loop:
	for err == nil {
		select {
		case <-ticker.C:
			if err = psc.Ping(""); err != nil {
				break loop
			}
		case <-ctx.Done():
			break loop
		case err := <-done:
			return err
		}
	}
	psc.Unsubscribe()
	return <-done
}
