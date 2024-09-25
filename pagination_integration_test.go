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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// using aircraft
var (
	brandAircraft    = "Boeing"
	categoryAircraft = "Narrow Body"
	aircraft         = NewMongoItem(Aircraft{
		Brand:    brandAircraft,
		Category: categoryAircraft,
		Engine: []Engine{
			{
				Provider: "Rolls Royce",
				Power:    1000,
			},
			{
				Provider: "Rolls Royce",
				Power:    1000,
			},
		},
	})

	itemKeyFormatAircraft        = "aircraft:%s"
	paginationKeyFormatAircraft  = "aircraft:brands:%s:type:%s"
	paginationParametersAircraft = []string{"Boeing", "Narrow Body"}
	paginationFilterAircraft     = bson.A{
		bson.D{{"aircraft", brandAircraft}},
		bson.D{{"categoryAircraft", categoryAircraft}},
	}
	keyAircraft = concatKey(paginationKeyFormatAircraft, paginationParametersAircraft)
)

type Engine struct {
	Provider string
	Power    int64
}

type Aircraft struct {
	*Item
	*MongoItem
	Ranking  int64    `bson:"ranking"`
	Brand    string   `bson:"brandAircraft"`
	Category string   `bson:"categoryAircraft"`
	Engine   []Engine `bson:"engine"`
}

type AircraftCustomDescend struct {
	*Item
	*MongoItem
	Ranking  int64    `bson:"ranking" sorting:"descending"`
	Brand    string   `bson:"brandAircraft"`
	Category string   `bson:"categoryAircraft"`
	Engine   []Engine `bson:"engine"`
}

type AircraftDefaultAscend struct {
	*Item `sorting:"ascending"`
	*MongoItem
	Ranking  int64    `bson:"ranking"`
	Brand    string   `bson:"brandAircraft"`
	Category string   `bson:"categoryAircraft"`
	Engine   []Engine `bson:"engine"`
}

type AircraftCustomAscend struct {
	*Item
	*MongoItem
	Ranking  int64    `bson:"ranking" sorting:"ascending"`
	Brand    string   `bson:"brandAircraft"`
	Category string   `bson:"categoryAircraft"`
	Engine   []Engine `bson:"engine"`
}

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

func TestAddItemIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	dummyAircraft := Aircraft{
		Brand:    "Volkswagen",
		Category: "SUV",
		Engine: []Engine{
			{
				Provider: "Rolls Royce",
				Power:    1000,
			},
			{
				Provider: "Rolls Royce",
				Power:    1000,
			},
		},
	}

	dummyAircraft = NewMongoItem(dummyAircraft)
	assert.NotNil(t, dummyAircraft)

	t.Run("add item without sorted-set", func(t *testing.T) {
		collection := connectMongo()
		mongo := Mongo[Aircraft](logger, collection)

		pagination := Pagination[Aircraft](
			paginationKeyFormatAircraft,
			itemKeyFormatAircraft,
			logger,
			connectRedis(),
		)
		pagination.WithMongo(mongo, paginationFilterAircraft)

		errorAddItem := pagination.AddItem(paginationParametersAircraft, dummyAircraft)
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
		key := concatKey(paginationKeyFormatAircraft, paginationParametersAircraft)
		zadd := redisClient.ZAdd(
			context.TODO(),
			key,
			member,
		)
		assert.Nil(t, zadd.Err())

		// test starting point
		mongo := Mongo[Aircraft](logger, collection)
		pagination := Pagination[Aircraft](
			paginationKeyFormatAircraft,
			itemKeyFormatAircraft,
			logger,
			connectRedis(),
		)
		pagination.WithMongo(mongo, paginationFilterAircraft)

		errorAddItem := pagination.AddItem(paginationParametersAircraft, dummyAircraft)
		assert.Nil(t, errorAddItem)

		// collection.Database().Drop(context.TODO())
	})

}

func TestFetchOneIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Run("fetch success", func(t *testing.T) {
		pagination := Pagination[Aircraft](
			paginationKeyFormatAircraft,
			itemKeyFormatAircraft,
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
		fmt.Println(result.Engine)
	})
}

func TestSeedOneIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Run("seed success", func(t *testing.T) {
		mongo := Mongo[Aircraft](logger, connectMongo())
		pagination := Pagination[Aircraft](
			paginationKeyFormatAircraft,
			itemKeyFormatAircraft,
			logger,
			connectRedis(),
		)
		pagination.WithMongo(mongo, paginationFilterAircraft)

		car, errorSeed := pagination.SeedOne("rivaOehekZC0BIHN")
		assert.Nil(t, errorSeed)
		assert.NotNil(t, car)

		fmt.Println(car)
	})
}
