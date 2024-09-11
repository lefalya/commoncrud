package commonpagination

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func ConnectMongo() *mongo.Collection {
	URI := os.Getenv("MONGO_HOST")
	database := os.Getenv("MONGO_DATABASE")
	collection := os.Getenv("MONGO_COLLECTION")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	client, errConnect := mongo.Connect(ctx, options.Client().ApplyURI(URI))

	if errConnect != nil {
		panic(errConnect)
	}

	if errPing := client.Ping(ctx, readpref.Primary()); errPing != nil {
		panic(errPing)
	}

	return client.Database(database).Collection(collection)
}

func connectRedis() redis.UniversalClient {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		log.Fatal("REDIS_HOST environment variable not set")
	}

	client := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    []string{redisHost},
		Password: os.Getenv("REDIS_PASS"),
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func TestCreateMongoDB(t *testing.T) {}
