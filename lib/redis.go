package lib

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/redis/go-redis/v9"
	"redmon/models"
)

func RedisClientFromConfig(config *models.Config) (*redis.Client, error) {
	tlsConf := &tls.Config{}
	if !config.UseTLS {
		tlsConf = nil
	}

	addr := fmt.Sprintf("%s:%d", config.RedisHost, config.RedisPort)
	client := redis.NewClient(&redis.Options{
		Addr:       addr,
		ClientName: "Redmon",
		Username:   config.RedisUser,
		Password:   config.RedisPassword,
		DB:         config.RedisDB,
		TLSConfig:  tlsConf,
	})

	err := client.Ping(context.Background()).Err()
	return client, err
}
