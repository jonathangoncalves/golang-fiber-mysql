package initializers

import (
	"context"
	"github.com/go-redis/redis/v8"

	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/wpcodevo/golang-fiber-mysql/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB  *gorm.DB
	ctx = context.Background()
	rdb *redis.Client
)

func ConnectDB(config *Config) {
	var err error
	// dsn := fmt.Sprintf("user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.DBUserName, config.DBUserPassword, config.DBHost, config.DBPort, config.DBName)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to the Database! \n", err.Error())
		os.Exit(1)
	}

	DB.Logger = logger.Default.LogMode(logger.Info)

	log.Println("Running Migrations")
	DB.AutoMigrate(&models.Note{})

	log.Println("🚀 Connected Successfully to the Database")

	initRedis(1)
}

func initRedis(selectDB ...int) {
	// Connect to Redis

	var redisHost = os.Getenv("REDIS_HOST")
	var redisPassword = os.Getenv("REDIS_PASSWORD")

	rdb = redis.NewClient(&redis.Options{
		Addr:     redisHost,     // Redis server address
		Password: redisPassword, // Redis password, if any
		DB:       selectDB[0],   // Redis database number
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic("Failed to connect to Redis")
	}

}

func SetCache(key string, value interface{}, expiration int) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return rdb.Set(ctx, key, data, time.Duration(expiration)*time.Second).Err()
}

func GetCache(key string, dest interface{}) error {
	data, err := rdb.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}
