package lib

import (
	"context"
	"crypto/tls"
	"net"
	"strconv"

	"github.com/redis/go-redis/v9"

	"redscout/models"
)

func RedisClientFromConfig(config *models.Config) (*redis.Client, error) {
	var tlsConf *tls.Config
	if config.UseTLS {
		tlsConf = &tls.Config{}
	}

	client := redis.NewClient(&redis.Options{
		Addr:       net.JoinHostPort(config.RedisHost, strconv.Itoa(config.RedisPort)),
		ClientName: "redscout",
		Username:   config.RedisUser,
		Password:   config.RedisPassword,
		DB:         config.RedisDB,
		TLSConfig:  tlsConf,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}
