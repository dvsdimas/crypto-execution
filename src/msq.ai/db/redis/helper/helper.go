package helper

import (
	"github.com/go-errors/errors"
	"github.com/go-redis/redis"
)

func GetRedisClient(url string, password *string, db *int) (*redis.Client, error) {

	var dbIdx = 0

	if db != nil {
		if *db < 0 {
			return nil, errors.New("db index cannot be less than zero")
		}

		dbIdx = *db
	}

	var pwd = ""

	if password != nil {
		pwd = *password
	}

	client := redis.NewClient(&redis.Options{Addr: url, Password: pwd, DB: dbIdx})

	_, err := client.Ping().Result()

	if err != nil {
		return nil, err
	}

	return client, nil
}

func CloseRedisClient(client *redis.Client) {

}
