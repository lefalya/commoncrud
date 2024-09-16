package commoncrud

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func connectMongo() *mongo.Collection {
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

func TestIntegrationAddItem(t *testing.T) {
	dummyCar := Car{
		Brand:    "Volkswagen",
		Category: "SUV",
		Seating: []Seater{
			{
				Material:  "Leather",
				Occupancy: 2,
			},
			{
				Material:  "Leather",
				Occupancy: 3,
			},
			{
				Material:  "Leather",
				Occupancy: 2,
			},
		},
	}

	dummyCar = NewMongoItem(dummyCar)
	assert.NotNil(t, dummyCar)

	t.Run("add item without sorted-set", func(t *testing.T) {
		collection := connectMongo()
		mongo := Mongo[Car](logger, collection)

		pagination := Pagination[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			connectRedis(),
		)
		pagination.WithMongo(mongo, paginationFilter)

		errorAddItem := pagination.AddItem(paginationParameters, dummyCar)
		assert.Nil(t, errorAddItem)

		collection.Database().Drop(context.TODO())
	})

	t.Run("add item with sorted-set added", func(t *testing.T) {
		collection := connectMongo()

		// creating dummy sorted set
		redisClient := connectRedis()
		member := redis.Z{
			Score:  float64(time.Now().Unix()),
			Member: RandId(),
		}
		key := concatKey(paginationKeyFormat, paginationParameters)
		zadd := redisClient.ZAdd(
			context.TODO(),
			key,
			member,
		)
		assert.Nil(t, zadd.Err())

		// test starting point
		mongo := Mongo[Car](logger, collection)
		pagination := Pagination[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			connectRedis(),
		)
		pagination.WithMongo(mongo, paginationFilter)

		errorAddItem := pagination.AddItem(paginationParameters, dummyCar)
		assert.Nil(t, errorAddItem)

		// collection.Database().Drop(context.TODO())
	})

}

func TestIntegrationFetchOne(t *testing.T) {
	t.Run("fetch success", func(t *testing.T) {
		pagination := Pagination[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			connectRedis(),
		)

		result, errorFetch := pagination.FetchOne("7leQ0wvP0igqnBxs")
		assert.Nil(t, errorFetch)
		assert.NotNil(t, result)
		fmt.Println(result)
		fmt.Println(result.CreatedAt)
		fmt.Println(result.UUID)
		fmt.Println(result.RandId)
		fmt.Println(result.Seating)
	})
}

func TestIntegrationSeedOne(t *testing.T) {

	t.Run("seed success", func(t *testing.T) {
		mongo := Mongo[Car](logger, connectMongo())
		pagination := Pagination[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			connectRedis(),
		)
		pagination.WithMongo(mongo, paginationFilter)

		car, errorSeed := pagination.SeedOne("rivaOehekZC0BIHN")
		assert.Nil(t, errorSeed)
		assert.NotNil(t, car)

		fmt.Println(car)
	})
}
