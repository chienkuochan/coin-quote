package utils

import (
	"context"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
)

const healthCheckPeriod = time.Minute

var (
	DatePool    *redis.Pool
	LimitPool   *redis.Pool
	PublishPool *redis.Pool
)

func InitRedisPools() {
	DatePool = newRedisPool("data")
	LimitPool = newRedisPool("limit")
	PublishPool = newRedisPool("publish")
}

func CloseRedisPools() {
	DatePool.Close()
	LimitPool.Close()
	PublishPool.Close()
}

func GetRedisConn(redisName string) redis.Conn {
	if redisName == "data" {
		return DatePool.Get()
	} else if redisName == "limit" {
		return LimitPool.Get()
	} else if redisName == "publish" {
		return PublishPool.Get()
	}
	return nil
}

func newRedisPool(redisName string) *redis.Pool {
	config := getRedisConfig()
	capacity := config.GetInt(redisName+".pool", 10)
	maxCapacity := config.GetInt(redisName+".maxopen", 0)
	idleTimout := config.GetDuration(redisName+".timeout", "4m")
	maxConnLifetime := config.GetDuration(redisName+".life_time", "2m")
	network := config.Get(redisName+".network", "tcp")
	server := config.Get(redisName+".server", "localhost:6379")
	db := config.Get(redisName+".db", "")
	password := config.Get(redisName+".password", "")

	return &redis.Pool{
		MaxIdle:         capacity,
		MaxActive:       maxCapacity,
		IdleTimeout:     idleTimout,
		MaxConnLifetime: maxConnLifetime,
		Wait:            true,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial(network, server)
			if err != nil {
				log.Println("redis can't dial:" + err.Error())
				return nil, err
			}

			if password != "" {
				_, err := conn.Do("AUTH", password)
				if err != nil {
					log.Println("redis can't AUTH:" + err.Error())
					conn.Close()
					return nil, err
				}
			}

			if db != "" {
				_, err := conn.Do("SELECT", db)
				if err != nil {
					log.Println("redis can't SELECT:" + err.Error())
					conn.Close()
					return nil, err
				}
			}
			return conn, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				log.Println("redis can't ping, err:" + err.Error())
			}
			return err
		},
	}
}

func ListenPubSubChannels(ctx context.Context, onStart func() error, onMessage func(channel string, data []byte) error, channels ...string) (err error) {
	dataRedis := GetRedisConn("publish")
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

func PublishToPubSubChannels(channel string, message *[]byte) {
	dataRedis := GetRedisConn("publish")
	defer dataRedis.Close()
	dataRedis.Do("publish", channel, string(*message))
}
