package cache

import (
	"github.com/redis/rueidis"
	"github.com/sirupsen/logrus"
)

func NewRedis(address string, password string) rueidis.Client {
	client, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress: []string{address},
		Username:    "default",
		Password:    password,
	})
	if err != nil {
		logrus.WithError(err).Panicln("couldn't connect to redis")
	}
	return client
}
