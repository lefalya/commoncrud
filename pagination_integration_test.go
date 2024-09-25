package commoncrud

import (
	"go.mongodb.org/mongo-driver/bson"
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

/**
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
	t.Run("add item without sorted-set", func(t *testing.T) {
		dummyAircraft := NewMongoItem(Aircraft{
			Brand:    "Boeing",
			Category: "Narrow Body",
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

		collection := connectMongo()
		redisClient := connectRedis()
		mongo := Mongo[Aircraft](logger, collection)

		pagination := Pagination[Aircraft](
			paginationKeyFormatAircraft,
			itemKeyFormatAircraft,
			logger,
			redisClient,
		)
		pagination.WithMongo(mongo, paginationFilterAircraft)

		errorAddItem := pagination.AddItem(paginationParametersAircraft, dummyAircraft)
		assert.Nil(t, errorAddItem)

		collection.Database().Drop(context.TODO())
		redisClient.FlushAll(context.TODO())
	})
	t.Run("add item with no database specified", func(t *testing.T) {
		// creating dummy sorted set
		redisClient := connectRedis()
		member := redis.Z{
			Score:  float64(time.Now().Unix()),
			Member: RandId(),
		}
		key := concatKey(paginationKeyFormatAircraft, paginationParametersAircraft)
		zadd := redisClient.ZAdd(
			context.TODO(),
			key+descendingTrailing+"createdat",
			member,
		)
		assert.Nil(t, zadd.Err())
		time.Sleep(100 * time.Millisecond)

		dummyAircraft := NewMongoItem(Aircraft{
			Brand:    "Boeing",
			Category: "Narrow Body",
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

		pagination := Pagination[Aircraft](
			paginationKeyFormatAircraft,
			itemKeyFormatAircraft,
			logger,
			connectRedis(),
		)
		errorAddItem := pagination.AddItem(paginationParametersAircraft, dummyAircraft)
		assert.Nil(t, errorAddItem)

		redisClient.FlushAll(context.TODO())
	})
	t.Run("with descend sorting on default attribute", func(t *testing.T) {
		collection := connectMongo()

		// creating dummy sorted set
		redisClient := connectRedis()
		member := redis.Z{
			Score:  float64(time.Now().UnixMilli()),
			Member: RandId(),
		}
		key := concatKey(paginationKeyFormatAircraft, paginationParametersAircraft)
		zadd := redisClient.ZAdd(
			context.TODO(),
			key+descendingTrailing+"createdat",
			member,
		)
		assert.Nil(t, zadd.Err())
		time.Sleep(100 * time.Millisecond)

		dummyAircraft := NewMongoItem(Aircraft{
			Brand:    "Boeing",
			Category: "Narrow Body",
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

		members := redisClient.ZRevRange(
			context.TODO(),
			key+descendingTrailing+"createdat",
			0, -1,
		)
		assert.Nil(t, members.Err())
		assert.Equal(t, 2, len(members.Val()))
		assert.Equal(t, dummyAircraft.GetRandId(), members.Val()[0])

		collection.Database().Drop(context.TODO())
		redisClient.FlushAll(context.TODO())
	})
	t.Run("with descend sorting on custom attribute", func(t *testing.T) {
		collection := connectMongo()

		// creating dummy sorted set
		redisClient := connectRedis()
		member := redis.Z{
			Score:  float64(10),
			Member: RandId(),
		}
		key := concatKey(paginationKeyFormatAircraft, paginationParametersAircraft)
		zadd := redisClient.ZAdd(
			context.TODO(),
			key+descendingTrailing+"ranking",
			member,
		)
		assert.Nil(t, zadd.Err())
		time.Sleep(100 * time.Millisecond)

		dummyAircraft1 := NewMongoItem(AircraftCustomDescend{
			Brand:    "Boeing",
			Category: "Narrow Body",
			Ranking:  11,
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
		dummyAircraft2 := NewMongoItem(AircraftCustomDescend{
			Brand:    "Boeing",
			Category: "Narrow Body",
			Ranking:  12,
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

		// test starting point
		mongo := Mongo[AircraftCustomDescend](logger, collection)
		pagination := Pagination[AircraftCustomDescend](
			paginationKeyFormatAircraft,
			itemKeyFormatAircraft,
			logger,
			redisClient,
		)
		pagination.WithMongo(mongo, paginationFilterAircraft)

		errorAddItem := pagination.AddItem(paginationParametersAircraft, dummyAircraft1)
		assert.Nil(t, errorAddItem)
		errorAddItem = pagination.AddItem(paginationParametersAircraft, dummyAircraft2)
		assert.Nil(t, errorAddItem)

		members := redisClient.ZRevRange(
			context.TODO(),
			key+descendingTrailing+"ranking",
			0, -1,
		)
		assert.Nil(t, members.Err())
		assert.Equal(t, 3, len(members.Val()))
		assert.Equal(t, dummyAircraft2.GetRandId(), members.Val()[0])
		assert.Equal(t, dummyAircraft1.GetRandId(), members.Val()[1])

		collection.Database().Drop(context.TODO())
		redisClient.FlushAll(context.TODO())
	})
	t.Run("with ascend sorting on default attribute", func(t *testing.T) {
		collection := connectMongo()

		// creating dummy sorted set
		redisClient := connectRedis()
		member := redis.Z{
			Score:  float64(10),
			Member: RandId(),
		}
		key := concatKey(paginationKeyFormatAircraft, paginationParametersAircraft)
		zadd := redisClient.ZAdd(
			context.TODO(),
			key+ascendingTrailing+"createdat",
			member,
		)
		assert.Nil(t, zadd.Err())
		time.Sleep(100 * time.Millisecond)

		dummyAircraft1 := NewMongoItem(AircraftDefaultAscend{
			Brand:    "Boeing",
			Category: "Narrow Body",
			Ranking:  11,
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
		time.Sleep(100 * time.Millisecond)

		dummyAircraft2 := NewMongoItem(AircraftDefaultAscend{
			Brand:    "Boeing",
			Category: "Narrow Body",
			Ranking:  12,
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

		// test starting point
		mongo := Mongo[AircraftDefaultAscend](logger, collection)
		pagination := Pagination[AircraftDefaultAscend](
			paginationKeyFormatAircraft,
			itemKeyFormatAircraft,
			logger,
			redisClient,
		)
		pagination.WithMongo(mongo, paginationFilterAircraft)

		errorAddItem := pagination.AddItem(paginationParametersAircraft, dummyAircraft1)
		assert.Nil(t, errorAddItem)
		errorAddItem = pagination.AddItem(paginationParametersAircraft, dummyAircraft2)
		assert.Nil(t, errorAddItem)

		members := redisClient.ZRange(
			context.TODO(),
			key+ascendingTrailing+"createdat",
			0, -1,
		)
		assert.Nil(t, members.Err())
		assert.Equal(t, 3, len(members.Val()))
		assert.Equal(t, dummyAircraft2.GetRandId(), members.Val()[2])
		assert.Equal(t, dummyAircraft1.GetRandId(), members.Val()[1])

		collection.Database().Drop(context.TODO())
		redisClient.FlushAll(context.TODO())
	})
	t.Run("with ascend sorting on custom attribute", func(t *testing.T) {
		collection := connectMongo()

		// creating dummy sorted set
		redisClient := connectRedis()
		member := redis.Z{
			Score:  float64(10),
			Member: RandId(),
		}
		key := concatKey(paginationKeyFormatAircraft, paginationParametersAircraft)
		zadd := redisClient.ZAdd(
			context.TODO(),
			key+ascendingTrailing+"ranking",
			member,
		)
		assert.Nil(t, zadd.Err())
		time.Sleep(100 * time.Millisecond)

		dummyAircraft1 := NewMongoItem(AircraftCustomAscend{
			Brand:    "Boeing",
			Category: "Narrow Body",
			Ranking:  11,
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
		time.Sleep(100 * time.Millisecond)

		dummyAircraft2 := NewMongoItem(AircraftCustomAscend{
			Brand:    "Boeing",
			Category: "Narrow Body",
			Ranking:  12,
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

		// test starting point
		mongo := Mongo[AircraftCustomAscend](logger, collection)
		pagination := Pagination[AircraftCustomAscend](
			paginationKeyFormatAircraft,
			itemKeyFormatAircraft,
			logger,
			redisClient,
		)
		pagination.WithMongo(mongo, paginationFilterAircraft)

		errorAddItem := pagination.AddItem(paginationParametersAircraft, dummyAircraft1)
		assert.Nil(t, errorAddItem)
		errorAddItem = pagination.AddItem(paginationParametersAircraft, dummyAircraft2)
		assert.Nil(t, errorAddItem)

		members := redisClient.ZRange(
			context.TODO(),
			key+ascendingTrailing+"ranking",
			0, -1,
		)
		assert.Nil(t, members.Err())
		assert.Equal(t, 3, len(members.Val()))
		assert.Equal(t, dummyAircraft2.GetRandId(), members.Val()[2])
		assert.Equal(t, dummyAircraft1.GetRandId(), members.Val()[1])

		collection.Database().Drop(context.TODO())
		redisClient.FlushAll(context.TODO())
	})
}

func TestUpdateItemIntegration(t *testing.T) {

}

func TestRemoveItemIntegration(t *testing.T) {
	t.Run("successfully remove item", func(t *testing.T) {
		redisClient := connectRedis()
		member := redis.Z{
			Score:  float64(time.Now().UnixMilli()),
			Member: RandId(),
		}
		key := concatKey(paginationKeyFormatAircraft, paginationParametersAircraft)
		zadd := redisClient.ZAdd(
			context.TODO(),
			key+descendingTrailing+"createdat",
			member,
		)
		assert.Nil(t, zadd.Err())
		time.Sleep(100 * time.Millisecond)

		dummyAircraft := NewMongoItem(Aircraft{
			Brand:    "Boeing",
			Category: "Narrow Body",
			Ranking:  10,
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

		collection := connectMongo()
		mongo := Mongo[Aircraft](logger, collection)

		pagination := Pagination[Aircraft](
			paginationKeyFormatAircraft,
			itemKeyFormatAircraft,
			logger,
			redisClient,
		)
		pagination.WithMongo(mongo, paginationFilterAircraft)

		errorAddItem := pagination.AddItem(paginationParametersAircraft, dummyAircraft)
		assert.Nil(t, errorAddItem)

		errorDeleteItem := pagination.RemoveItem(paginationParametersAircraft, dummyAircraft)
		assert.Nil(t, errorDeleteItem)

		collection.Database().Drop(context.TODO())
		redisClient.FlushAll(context.TODO())
	})
	t.Run("successfully remove item with no database specified", func(t *testing.T) {
		redisClient := connectRedis()
		member := redis.Z{
			Score:  float64(time.Now().UnixMilli()),
			Member: RandId(),
		}
		key := concatKey(paginationKeyFormatAircraft, paginationParametersAircraft)
		zadd := redisClient.ZAdd(
			context.TODO(),
			key+descendingTrailing+"createdat",
			member,
		)
		assert.Nil(t, zadd.Err())
		time.Sleep(100 * time.Millisecond)

		dummyAircraft := NewMongoItem(Aircraft{
			Brand:    "Boeing",
			Category: "Narrow Body",
			Ranking:  10,
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

		pagination := Pagination[Aircraft](
			paginationKeyFormatAircraft,
			itemKeyFormatAircraft,
			logger,
			redisClient,
		)

		errorAddItem := pagination.AddItem(paginationParametersAircraft, dummyAircraft)
		assert.Nil(t, errorAddItem)

		errorDeleteItem := pagination.RemoveItem(paginationParametersAircraft, dummyAircraft)
		assert.Nil(t, errorDeleteItem)

		redisClient.FlushAll(context.TODO())
	})
	t.Run("with descend sorting order on custom attribute", func(t *testing.T) {
		redisClient := connectRedis()
		member := redis.Z{
			Score:  float64(time.Now().UnixMilli()),
			Member: RandId(),
		}
		key := concatKey(paginationKeyFormatAircraft, paginationParametersAircraft)
		zadd := redisClient.ZAdd(
			context.TODO(),
			key+descendingTrailing+"ranking",
			member,
		)
		assert.Nil(t, zadd.Err())
		time.Sleep(100 * time.Millisecond)

		dummyAircraft := NewMongoItem(AircraftCustomDescend{
			Brand:    "Boeing",
			Category: "Narrow Body",
			Ranking:  10,
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

		collection := connectMongo()
		mongo := Mongo[AircraftCustomDescend](logger, collection)

		pagination := Pagination[AircraftCustomDescend](
			paginationKeyFormatAircraft,
			itemKeyFormatAircraft,
			logger,
			redisClient,
		)
		pagination.WithMongo(mongo, paginationFilterAircraft)

		errorAddItem := pagination.AddItem(paginationParametersAircraft, dummyAircraft)
		assert.Nil(t, errorAddItem)

		errorDeleteItem := pagination.RemoveItem(paginationParametersAircraft, dummyAircraft)
		assert.Nil(t, errorDeleteItem)

		collection.Database().Drop(context.TODO())
		redisClient.FlushAll(context.TODO())
	})
	t.Run("with ascend sorting order on default attribute", func(t *testing.T) {
		redisClient := connectRedis()
		member := redis.Z{
			Score:  float64(time.Now().UnixMilli()),
			Member: RandId(),
		}
		key := concatKey(paginationKeyFormatAircraft, paginationParametersAircraft)
		zadd := redisClient.ZAdd(
			context.TODO(),
			key+ascendingTrailing+"createdat",
			member,
		)
		assert.Nil(t, zadd.Err())
		time.Sleep(100 * time.Millisecond)

		dummyAircraft := NewMongoItem(AircraftDefaultAscend{
			Brand:    "Boeing",
			Category: "Narrow Body",
			Ranking:  10,
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

		collection := connectMongo()
		mongo := Mongo[AircraftDefaultAscend](logger, collection)

		pagination := Pagination[AircraftDefaultAscend](
			paginationKeyFormatAircraft,
			itemKeyFormatAircraft,
			logger,
			redisClient,
		)
		pagination.WithMongo(mongo, paginationFilterAircraft)

		errorAddItem := pagination.AddItem(paginationParametersAircraft, dummyAircraft)
		assert.Nil(t, errorAddItem)

		errorDeleteItem := pagination.RemoveItem(paginationParametersAircraft, dummyAircraft)
		assert.Nil(t, errorDeleteItem)

		collection.Database().Drop(context.TODO())
		redisClient.FlushAll(context.TODO())
	})
	t.Run("with ascend sorting order on custom attribute", func(t *testing.T) {

	})
}

func TestTotalItemOnCacheIntergration(t *testing.T) {

	t.Run("succesfully get total items from cache", func(t *testing.T) {

	})
	t.Run("with descend sorting order on custom attribute", func(t *testing.T) {

	})
	t.Run("with ascend sorting order on default attribute", func(t *testing.T) {

	})
	t.Run("with ascend sorting order on custom attribute", func(t *testing.T) {

	})
}

func TestFetchOneIntegration(t *testing.T) {
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
**/
