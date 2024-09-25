package commoncrud

import (
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/golang/mock/gomock"
	"github.com/lefalya/commoncrud/interfaces"
	mock_interfaces "github.com/lefalya/commoncrud/mocks"
	"github.com/lefalya/commoncrud/types"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	brand    = "Volkswagen"
	category = "SUV"
	car      = NewMongoItem(Car{
		Brand:    brand,
		Category: category,
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
	})

	itemKeyFormat        = "car:%s"
	paginationKeyFormat  = "car:brands:%s:type:%s"
	paginationParameters = []string{"Volkswagen", "SUV"}
	paginationFilter     = bson.A{
		bson.D{{"car", brand}},
		bson.D{{"category", category}},
	}
	key = concatKey(paginationKeyFormat, paginationParameters)
)

func initTestPaginationType[T interfaces.Item](
	pagKeyFormat string,
	itemKeyFormat string,
	logger *slog.Logger,
	redisClient redis.UniversalClient,
	itemCache interfaces.ItemCache[T],
) *PaginationType[T] {
	pagination := PaginationType[T]{
		pagKeyFormat:  pagKeyFormat,
		itemKeyFormat: itemKeyFormat,
		logger:        logger,
		redisClient:   redisClient,
		itemCache:     itemCache,
	}

	sortOpt := SetSorting[T]()
	if sortOpt != nil {
		pagination.sorting = sortOpt
	}

	return &pagination
}

type Seater struct {
	Material  string
	Occupancy int64
}

type Car struct {
	*Item
	*MongoItem
	Ranking  int64    `bson:"ranking"`
	Brand    string   `bson:"brand"`
	Category string   `bson:"category"`
	Seating  []Seater `bson:"seating"`
}

type CarCustomDescend struct {
	*Item
	*MongoItem
	Ranking  int64    `bson:"ranking" sorting:"descending"`
	Brand    string   `bson:"brand"`
	Category string   `bson:"category"`
	Seating  []Seater `bson:"seating"`
}

type CarDefaultAscend struct {
	*Item `sorting:"ascending"`
	*MongoItem
	Ranking  int64    `bson:"ranking"`
	Brand    string   `bson:"brand"`
	Category string   `bson:"category"`
	Seating  []Seater `bson:"seating"`
}

type CarCustomAscend struct {
	*Item
	*MongoItem
	Ranking  int64    `bson:"ranking" sorting:"ascending"`
	Brand    string   `bson:"brand"`
	Category string   `bson:"category"`
	Seating  []Seater `bson:"seating"`
}

func TestInjectPagination(t *testing.T) {
	type Injected[T interfaces.Item] struct {
		pagination interfaces.Pagination[T]
	}

	pagination := Pagination[Car]("", "", nil, nil)
	injected := Injected[Car]{
		pagination: pagination,
	}

	assert.NotNil(t, injected)
}

func TestInitPagiantion(t *testing.T) {
	t.Run("init pagination with sorting", func(t *testing.T) {
		pagination := Pagination[CarCustomDescend]("", "", nil, nil)
		assert.NotNil(t, pagination)
		assert.Equal(t, "ranking", pagination.sorting.attribute)
		assert.Equal(t, descending, pagination.sorting.direction)
	})
	t.Run("init pagination createdAt ascending", func(t *testing.T) {
		pagination := Pagination[CarDefaultAscend]("", "", nil, nil)
		assert.NotNil(t, pagination)
		assert.Equal(t, "createdat", pagination.sorting.attribute)
		assert.Equal(t, ascending, pagination.sorting.direction)
	})
}

func TestAddItem(t *testing.T) {
	t.Run("successfully add item without addition to sorted set", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mongoMock := mock_interfaces.NewMockMongo[Car](ctrl)
		mongoMock.EXPECT().SetPaginationFilter(nil)
		mongoMock.EXPECT().Create(car).Return(nil)

		redis, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + descendingTrailing + "createdat").SetVal(0)

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redis,
			nil,
		)
		pagination.WithMongo(mongoMock, nil)

		errorAddItem := pagination.AddItem(paginationParameters, car)
		assert.Nil(t, errorAddItem)
	})
	t.Run("successfully add item with addition to sorted set", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mongoMock := mock_interfaces.NewMockMongo[Car](ctrl)
		mongoMock.EXPECT().SetPaginationFilter(nil)
		mongoMock.EXPECT().Create(car).Return(nil)

		itemCacheMock := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCacheMock.EXPECT().Set(car).Return(nil)

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + descendingTrailing + "createdat").SetVal(3)
		expectedZMember := redis.Z{
			Score:  float64(car.GetCreatedAt().Unix()),
			Member: car.GetRandId(),
		}
		mockedRedis.ExpectZAdd(key+descendingTrailing+"createdat", expectedZMember).SetVal(1)
		mockedRedis.ExpectExpire(key+descendingTrailing+"createdat", SORTED_SET_TTL).SetVal(true)

		pagination := initTestPaginationType(
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)
		pagination.WithMongo(mongoMock, nil)
		assert.Nil(t, pagination.sorting)

		errorAddItem := pagination.AddItem(paginationParameters, car)
		assert.Nil(t, errorAddItem)
	})
	t.Run("successfully add item with no database specified", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		itemCacheMock := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCacheMock.EXPECT().Set(car).Return(nil)

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + descendingTrailing + "createdat").SetVal(3)
		expectedZMember := redis.Z{
			Score:  float64(car.GetCreatedAt().Unix()),
			Member: car.GetRandId(),
		}
		mockedRedis.ExpectZAdd(key+descendingTrailing+"createdat", expectedZMember).SetVal(1)
		mockedRedis.ExpectExpire(key+descendingTrailing+"createdat", SORTED_SET_TTL).SetVal(true)

		pagination := initTestPaginationType(
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)

		errorAddItem := pagination.AddItem(paginationParameters, car)
		assert.Nil(t, errorAddItem)
	})
	t.Run("with descend sorting on custom attribute", func(t *testing.T) {
		car := NewMongoItem(CarCustomDescend{
			Brand:    brand,
			Category: category,
			Ranking:  12,
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
		})

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mongoMock := mock_interfaces.NewMockMongo[CarCustomDescend](ctrl)
		mongoMock.EXPECT().SetPaginationFilter(nil)
		mongoMock.EXPECT().Create(car).Return(nil)

		itemCacheMock := mock_interfaces.NewMockItemCache[CarCustomDescend](ctrl)
		itemCacheMock.EXPECT().Set(car).Return(nil)

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + descendingTrailing + "ranking").SetVal(3)
		expectedZMember := redis.Z{
			Score:  float64(car.Ranking),
			Member: car.GetRandId(),
		}
		mockedRedis.ExpectZAdd(key+descendingTrailing+"ranking", expectedZMember).SetVal(1)
		mockedRedis.ExpectExpire(key+descendingTrailing+"ranking", SORTED_SET_TTL).SetVal(true)

		pagination := initTestPaginationType(
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)
		pagination.WithMongo(mongoMock, nil)

		errorAddItem := pagination.AddItem(paginationParameters, car)
		assert.Nil(t, errorAddItem)
		assert.Equal(t, "ranking", pagination.sorting.attribute)
		assert.Equal(t, descending, pagination.sorting.direction)
	})
	t.Run("with ascend sorting on default attribute", func(t *testing.T) {
		car := NewMongoItem(CarDefaultAscend{
			Brand:    brand,
			Category: category,
			Ranking:  12,
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
		})

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mongoMock := mock_interfaces.NewMockMongo[CarDefaultAscend](ctrl)
		mongoMock.EXPECT().SetPaginationFilter(nil)
		mongoMock.EXPECT().Create(car).Return(nil)

		itemCacheMock := mock_interfaces.NewMockItemCache[CarDefaultAscend](ctrl)
		itemCacheMock.EXPECT().Set(car).Return(nil)

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + ascendingTrailing + "createdat").SetVal(3)
		expectedZMember := redis.Z{
			Score:  float64(car.GetCreatedAt().Unix()),
			Member: car.GetRandId(),
		}
		mockedRedis.ExpectZAdd(key+ascendingTrailing+"createdat", expectedZMember).SetVal(1)
		mockedRedis.ExpectExpire(key+ascendingTrailing+"createdat", SORTED_SET_TTL).SetVal(true)

		pagination := initTestPaginationType(
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)
		pagination.WithMongo(mongoMock, nil)

		errorAddItem := pagination.AddItem(paginationParameters, car)
		assert.Nil(t, errorAddItem)
		assert.Equal(t, "createdat", pagination.sorting.attribute)
		assert.Equal(t, ascending, pagination.sorting.direction)
	})
	t.Run("with descend sorting on default attribute", func(t *testing.T) {
		car := NewMongoItem(Car{
			Brand:    brand,
			Category: category,
			Ranking:  12,
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
		})

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mongoMock := mock_interfaces.NewMockMongo[Car](ctrl)
		mongoMock.EXPECT().SetPaginationFilter(nil)
		mongoMock.EXPECT().Create(car).Return(nil)

		itemCacheMock := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCacheMock.EXPECT().Set(car).Return(nil)

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + descendingTrailing + "createdat").SetVal(3)
		expectedZMember := redis.Z{
			Score:  float64(car.GetCreatedAt().Unix()),
			Member: car.GetRandId(),
		}
		mockedRedis.ExpectZAdd(key+descendingTrailing+"createdat", expectedZMember).SetVal(1)
		mockedRedis.ExpectExpire(key+descendingTrailing+"createdat", SORTED_SET_TTL).SetVal(true)

		pagination := initTestPaginationType(
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)
		pagination.WithMongo(mongoMock, nil)
		assert.Nil(t, pagination.sorting)

		errorAddItem := pagination.AddItem(paginationParameters, car)
		assert.Nil(t, errorAddItem)
	})
	t.Run("with ascend sorting on custom attribute", func(t *testing.T) {
		car := NewMongoItem(CarCustomAscend{
			Brand:    brand,
			Category: category,
			Ranking:  12,
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
		})

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mongoMock := mock_interfaces.NewMockMongo[CarCustomAscend](ctrl)
		mongoMock.EXPECT().SetPaginationFilter(nil)
		mongoMock.EXPECT().Create(car).Return(nil)

		itemCacheMock := mock_interfaces.NewMockItemCache[CarCustomAscend](ctrl)
		itemCacheMock.EXPECT().Set(car).Return(nil)

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + ascendingTrailing + "ranking").SetVal(3)
		expectedZMember := redis.Z{
			Score:  float64(car.Ranking),
			Member: car.GetRandId(),
		}
		mockedRedis.ExpectZAdd(key+ascendingTrailing+"ranking", expectedZMember).SetVal(1)
		mockedRedis.ExpectExpire(key+ascendingTrailing+"ranking", SORTED_SET_TTL).SetVal(true)

		pagination := initTestPaginationType(
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)
		pagination.WithMongo(mongoMock, nil)

		errorAddItem := pagination.AddItem(paginationParameters, car)
		assert.Nil(t, errorAddItem)
		assert.Equal(t, "ranking", pagination.sorting.attribute)
		assert.Equal(t, ascending, pagination.sorting.direction)
	})
	t.Run("zcard fatal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mongoMock := mock_interfaces.NewMockMongo[Car](ctrl)
		mongoMock.EXPECT().Create(car).Return(nil)
		mongoMock.EXPECT().SetPaginationFilter(nil)

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + descendingTrailing + "createdat").SetErr(errors.New("fatal error: Redis connection lost"))

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			nil,
		)
		pagination.WithMongo(mongoMock, nil)

		errorAddItem := pagination.AddItem(paginationParameters, car)
		assert.NotNil(t, errorAddItem)
		assert.Equal(t, REDIS_FATAL_ERROR, errorAddItem.Err)

	})
	t.Run("itemcache set error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mongoMock := mock_interfaces.NewMockMongo[Car](ctrl)
		mongoMock.EXPECT().Create(car).Return(nil)
		mongoMock.EXPECT().SetPaginationFilter(nil)

		itemCacheMock := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCacheMock.EXPECT().Set(car).Return(&types.PaginationError{Err: REDIS_FATAL_ERROR})

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + descendingTrailing + "createdat").SetVal(3)

		pagination := initTestPaginationType(
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)
		pagination.WithMongo(mongoMock, nil)

		errorAddItem := pagination.AddItem(paginationParameters, car)
		assert.NotNil(t, errorAddItem)
		assert.Equal(t, REDIS_FATAL_ERROR, errorAddItem.Err)
	})
	t.Run("zadd fatal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mongoMock := mock_interfaces.NewMockMongo[Car](ctrl)
		mongoMock.EXPECT().Create(car).Return(nil)
		mongoMock.EXPECT().SetPaginationFilter(nil)

		itemCacheMock := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCacheMock.EXPECT().Set(car).Return(nil)

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + descendingTrailing + "createdat").SetVal(3)
		expectedZMember := redis.Z{
			Score:  float64(car.GetCreatedAt().Unix()),
			Member: car.GetRandId(),
		}
		mockedRedis.ExpectZAdd(key+descendingTrailing+"createdat", expectedZMember).SetErr(errors.New("fatal error: Redis connection lost"))

		pagination := initTestPaginationType(
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)
		pagination.WithMongo(mongoMock, nil)

		errorAddItem := pagination.AddItem(paginationParameters, car)
		assert.NotNil(t, errorAddItem)
		assert.Equal(t, REDIS_FATAL_ERROR, errorAddItem.Err)
	})
	t.Run("set expire fatal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mongoMock := mock_interfaces.NewMockMongo[Car](ctrl)
		mongoMock.EXPECT().Create(car).Return(nil)
		mongoMock.EXPECT().SetPaginationFilter(nil)

		itemCacheMock := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCacheMock.EXPECT().Set(car).Return(nil)

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + descendingTrailing + "createdat").SetVal(3)
		expectedZMember := redis.Z{
			Score:  float64(car.GetCreatedAt().Unix()),
			Member: car.GetRandId(),
		}
		mockedRedis.ExpectZAdd(key+descendingTrailing+"createdat", expectedZMember).SetErr(errors.New("fatal error: Redis connection lost"))

		pagination := initTestPaginationType(
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)
		pagination.WithMongo(mongoMock, nil)

		errorAddItem := pagination.AddItem(paginationParameters, car)
		assert.NotNil(t, errorAddItem)
		assert.Equal(t, REDIS_FATAL_ERROR, errorAddItem.Err)
	})
	t.Run("mongo create error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mongoMock := mock_interfaces.NewMockMongo[Car](ctrl)
		mongoMock.EXPECT().SetPaginationFilter(nil)
		mongoMock.EXPECT().Create(car).Return(&types.PaginationError{Err: MONGO_FATAL_ERROR})

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			nil,
			nil,
		)
		pagination.WithMongo(mongoMock, nil)

		errorAddItem := pagination.AddItem(paginationParameters, car)
		assert.NotNil(t, errorAddItem)
		assert.Equal(t, MONGO_FATAL_ERROR, errorAddItem.Err)
	})
}

func TestUpdateItem(t *testing.T) {
	t.Run("successfully update item", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mongoMock := mock_interfaces.NewMockMongo[Car](ctrl)
		mongoMock.EXPECT().SetPaginationFilter(nil)
		mongoMock.EXPECT().Update(car).Return(nil)

		itemCache := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCache.EXPECT().Set(car).Return(nil)

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			nil,
			itemCache,
		)
		pagination.WithMongo(mongoMock, nil)

		errorUpdateItem := pagination.UpdateItem(car)
		assert.Nil(t, errorUpdateItem)
	})
	t.Run("successfull update with no database specified", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		itemCache := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCache.EXPECT().Set(car).Return(nil)

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			nil,
			itemCache,
		)

		errorUpdateItem := pagination.UpdateItem(car)
		assert.Nil(t, errorUpdateItem)
	})
	t.Run("error mongo update", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mongoMock := mock_interfaces.NewMockMongo[Car](ctrl)
		mongoMock.EXPECT().SetPaginationFilter(nil)
		mongoMock.EXPECT().Update(car).Return(&types.PaginationError{Err: MONGO_FATAL_ERROR})

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			nil,
			nil,
		)
		pagination.WithMongo(mongoMock, nil)

		errorUpdateItem := pagination.UpdateItem(car)
		assert.NotNil(t, errorUpdateItem)
		assert.Equal(t, MONGO_FATAL_ERROR, errorUpdateItem.Err)
	})
	t.Run("error set itemcache", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mongoMock := mock_interfaces.NewMockMongo[Car](ctrl)
		mongoMock.EXPECT().SetPaginationFilter(nil)
		mongoMock.EXPECT().Update(car).Return(nil)

		itemCache := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCache.EXPECT().Set(car).Return(&types.PaginationError{Err: REDIS_FATAL_ERROR})

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			nil,
			itemCache,
		)
		pagination.WithMongo(mongoMock, nil)
		errorUpdateItem := pagination.UpdateItem(car)
		assert.NotNil(t, errorUpdateItem)
		assert.Equal(t, REDIS_FATAL_ERROR, errorUpdateItem.Err)
	})
}

func TestRemoveItem(t *testing.T) {
	t.Run("successfully remove item with no sorted set exits", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redis, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + descendingTrailing + "createdat").SetVal(0)

		mongoMock := mock_interfaces.NewMockMongo[Car](ctrl)
		mongoMock.EXPECT().SetPaginationFilter(nil)
		mongoMock.EXPECT().Delete(car).Return(nil)

		itemCache := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCache.EXPECT().Del(car).Return(nil)

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redis,
			itemCache,
		)
		pagination.WithMongo(mongoMock, nil)

		errorRemoveItem := pagination.RemoveItem(paginationParameters, car)
		assert.Nil(t, errorRemoveItem)
	})
	t.Run("remove item success with no database specified", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redis, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + descendingTrailing + "createdat").SetVal(20)
		mockedRedis.ExpectZRem(key+descendingTrailing+"createdat", car.GetRandId()).SetVal(1)

		itemCache := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCache.EXPECT().Del(car).Return(nil)

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redis,
			itemCache,
		)

		errorRemoveItem := pagination.RemoveItem(paginationParameters, car)
		assert.Nil(t, errorRemoveItem)
	})
	t.Run("with descend sorting on custom attribute", func(t *testing.T) {
		car := NewMongoItem(CarCustomDescend{
			Brand:    brand,
			Category: category,
			Ranking:  12,
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
		})

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redis, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + descendingTrailing + "ranking").SetVal(0)

		mongoMock := mock_interfaces.NewMockMongo[CarCustomDescend](ctrl)
		mongoMock.EXPECT().SetPaginationFilter(nil)
		mongoMock.EXPECT().Delete(car).Return(nil)

		itemCache := mock_interfaces.NewMockItemCache[CarCustomDescend](ctrl)
		itemCache.EXPECT().Del(car).Return(nil)

		pagination := initTestPaginationType[CarCustomDescend](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redis,
			itemCache,
		)
		pagination.WithMongo(mongoMock, nil)
		assert.NotNil(t, pagination.sorting)
		assert.Equal(t, "ranking", pagination.sorting.attribute)
		assert.Equal(t, 2, pagination.sorting.index)
		assert.Equal(t, descending, pagination.sorting.direction)

		errorRemoveItem := pagination.RemoveItem(paginationParameters, car)
		assert.Nil(t, errorRemoveItem)
	})
	t.Run("with ascend sorting on default attribute", func(t *testing.T) {
		car := NewMongoItem(CarDefaultAscend{
			Brand:    brand,
			Category: category,
			Ranking:  12,
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
		})

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redis, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + ascendingTrailing + "createdat").SetVal(0)

		mongoMock := mock_interfaces.NewMockMongo[CarDefaultAscend](ctrl)
		mongoMock.EXPECT().SetPaginationFilter(nil)
		mongoMock.EXPECT().Delete(car).Return(nil)

		itemCache := mock_interfaces.NewMockItemCache[CarDefaultAscend](ctrl)
		itemCache.EXPECT().Del(car).Return(nil)

		pagination := initTestPaginationType[CarDefaultAscend](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redis,
			itemCache,
		)
		pagination.WithMongo(mongoMock, nil)
		assert.Equal(t, "createdat", pagination.sorting.attribute)
		assert.Equal(t, 0, pagination.sorting.index)
		assert.Equal(t, ascending, pagination.sorting.direction)

		errorRemoveItem := pagination.RemoveItem(paginationParameters, car)
		assert.Nil(t, errorRemoveItem)
	})
	t.Run("with descend sorting on default attribute", func(t *testing.T) {
		car := NewMongoItem(Car{
			Brand:    brand,
			Category: category,
			Ranking:  12,
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
		})

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redis, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + descendingTrailing + "createdat").SetVal(0)

		mongoMock := mock_interfaces.NewMockMongo[Car](ctrl)
		mongoMock.EXPECT().SetPaginationFilter(nil)
		mongoMock.EXPECT().Delete(car).Return(nil)

		itemCache := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCache.EXPECT().Del(car).Return(nil)

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redis,
			itemCache,
		)
		pagination.WithMongo(mongoMock, nil)
		assert.Nil(t, pagination.sorting)

		errorRemoveItem := pagination.RemoveItem(paginationParameters, car)
		assert.Nil(t, errorRemoveItem)
	})
	t.Run("with ascend sorting on custom attribute", func(t *testing.T) {
		car := NewMongoItem(CarCustomAscend{
			Brand:    brand,
			Category: category,
			Ranking:  12,
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
		})

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redis, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + ascendingTrailing + "ranking").SetVal(0)

		mongoMock := mock_interfaces.NewMockMongo[CarCustomAscend](ctrl)
		mongoMock.EXPECT().SetPaginationFilter(nil)
		mongoMock.EXPECT().Delete(car).Return(nil)

		itemCache := mock_interfaces.NewMockItemCache[CarCustomAscend](ctrl)
		itemCache.EXPECT().Del(car).Return(nil)

		pagination := initTestPaginationType[CarCustomAscend](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redis,
			itemCache,
		)
		pagination.WithMongo(mongoMock, nil)
		assert.NotNil(t, pagination.sorting)
		assert.Equal(t, "ranking", pagination.sorting.attribute)
		assert.Equal(t, 2, pagination.sorting.index)
		assert.Equal(t, ascending, pagination.sorting.direction)

		errorRemoveItem := pagination.RemoveItem(paginationParameters, car)
		assert.Nil(t, errorRemoveItem)
	})
	t.Run("zcard fatal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + descendingTrailing + "createdat").SetErr(errors.New("fatal error: Redis connection lost"))

		itemCache := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCache.EXPECT().Del(car).Return(nil)

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCache,
		)

		errorRemoveItem := pagination.RemoveItem(paginationParameters, car)
		assert.NotNil(t, errorRemoveItem)
		assert.Equal(t, REDIS_FATAL_ERROR, errorRemoveItem.Err)
	})
	t.Run("zrem but item not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + descendingTrailing + "createdat").SetVal(3)
		mockedRedis.ExpectZRem(key+descendingTrailing+"createdat", car.GetRandId()).SetVal(0)

		itemCache := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCache.EXPECT().Del(car).Return(nil)

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCache,
		)

		errorRemoveItem := pagination.RemoveItem(paginationParameters, car)
		assert.Nil(t, errorRemoveItem)
	})
	t.Run("itemcache delete error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		itemCache := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCache.EXPECT().Del(car).Return(&types.PaginationError{Err: REDIS_FATAL_ERROR})

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			nil,
			itemCache,
		)

		errorRemoveItem := pagination.RemoveItem(paginationParameters, car)
		assert.NotNil(t, errorRemoveItem)
		assert.Equal(t, REDIS_FATAL_ERROR, errorRemoveItem.Err)
	})
	t.Run("mongo delete error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mongoMock := mock_interfaces.NewMockMongo[Car](ctrl)
		mongoMock.EXPECT().SetPaginationFilter(nil)
		mongoMock.EXPECT().Delete(car).Return(&types.PaginationError{Err: MONGO_FATAL_ERROR})

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			nil,
			nil,
		)
		pagination.WithMongo(mongoMock, nil)

		errorRemoveItem := pagination.RemoveItem(paginationParameters, car)
		assert.NotNil(t, errorRemoveItem)
		assert.Equal(t, MONGO_FATAL_ERROR, errorRemoveItem.Err)
	})
}

func TestTotalItemOnCache(t *testing.T) {
	t.Run("successfully get total items from cache", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + descendingTrailing + "createdat").SetVal(5)

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			nil,
		)
		assert.Nil(t, pagination.sorting)

		errorTotalItems := pagination.TotalItemOnCache(paginationParameters)

		assert.Nil(t, errorTotalItems)
	})
	t.Run("redis ZCard fatal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + descendingTrailing + "createdat").SetErr(errors.New("fatal error: Redis connection lost")) // Simulate a Redis connection error

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			nil,
		)
		assert.Nil(t, pagination.sorting)

		errorTotalItems := pagination.TotalItemOnCache(paginationParameters)

		assert.NotNil(t, errorTotalItems)
		assert.Equal(t, REDIS_FATAL_ERROR, errorTotalItems.Err)
	})
	t.Run("with descend sorting on custom attribute", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + descendingTrailing + "ranking").SetVal(5)

		pagination := initTestPaginationType[CarCustomDescend](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			nil,
		)
		assert.NotNil(t, pagination.sorting)
		assert.Equal(t, "ranking", pagination.sorting.attribute)
		assert.Equal(t, 2, pagination.sorting.index)
		assert.Equal(t, descending, pagination.sorting.direction)

		errorTotalItems := pagination.TotalItemOnCache(paginationParameters)

		assert.Nil(t, errorTotalItems)
	})
	t.Run("with ascend sorting on default attribute", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + descendingTrailing + "createdat").SetVal(5)

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			nil,
		)
		assert.Nil(t, pagination.sorting)

		errorTotalItems := pagination.TotalItemOnCache(paginationParameters)
		assert.Nil(t, errorTotalItems)
	})
	t.Run("with descend sorting on default attribute", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + descendingTrailing + "createdat").SetVal(5)

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			nil,
		)
		assert.Nil(t, pagination.sorting)

		errorTotalItems := pagination.TotalItemOnCache(paginationParameters)

		assert.Nil(t, errorTotalItems)
	})
	t.Run("with ascend sorting on custom attribute", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key + ascendingTrailing + "ranking").SetVal(5)

		pagination := initTestPaginationType[CarCustomAscend](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			nil,
		)
		assert.NotNil(t, pagination.sorting)
		assert.Equal(t, "ranking", pagination.sorting.attribute)
		assert.Equal(t, 2, pagination.sorting.index)
		assert.Equal(t, ascending, pagination.sorting.direction)

		errorTotalItems := pagination.TotalItemOnCache(paginationParameters)

		assert.Nil(t, errorTotalItems)
	})
}

func TestFetchAll(t *testing.T) {

	car1 := NewMongoItem(Car{Brand: "Toyota", Category: "SUV", Seating: []Seater{{Material: "Leather", Occupancy: 2}, {Material: "Leather", Occupancy: 3}, {Material: "Leather", Occupancy: 2}}})
	car2 := NewMongoItem(Car{Brand: "Honda", Category: "Sedan", Seating: []Seater{{Material: "Fabric", Occupancy: 2}, {Material: "Fabric", Occupancy: 3}}})
	car3 := NewMongoItem(Car{Brand: "Ford", Category: "Truck", Seating: []Seater{{Material: "Vinyl", Occupancy: 2}, {Material: "Vinyl", Occupancy: 2}}})
	car4 := NewMongoItem(Car{Brand: "BMW", Category: "Coupe", Seating: []Seater{{Material: "Leather", Occupancy: 2}, {Material: "Leather", Occupancy: 2}}})
	car5 := NewMongoItem(Car{Brand: "Tesla", Category: "Electric", Seating: []Seater{{Material: "Vegan Leather", Occupancy: 2}, {Material: "Vegan Leather", Occupancy: 3}}})

	carRandIds := []string{
		car1.GetRandId(),
		car2.GetRandId(),
		car3.GetRandId(),
		car4.GetRandId(),
		car5.GetRandId(),
	}

	t.Run("succesfully fetch all items", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZRevRange(key+descendingTrailing+"createdat", 0, -1).SetVal(carRandIds)

		itemCacheMock := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCacheMock.EXPECT().Get(car1.GetRandId()).Return(car1, nil)
		itemCacheMock.EXPECT().Get(car2.GetRandId()).Return(car2, nil)
		itemCacheMock.EXPECT().Get(car3.GetRandId()).Return(car3, nil)
		itemCacheMock.EXPECT().Get(car4.GetRandId()).Return(car4, nil)
		itemCacheMock.EXPECT().Get(car5.GetRandId()).Return(car5, nil)

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)

		fetchAll, errorFetchAll := pagination.FetchAll(
			paginationParameters,
			func(item Car, items *[]Car) {

				fmt.Println(item)

				*items = append(*items, item)
			})

		assert.Nil(t, errorFetchAll)
		assert.NotNil(t, fetchAll)
		assert.Equal(t, 5, len(fetchAll))
	})
	t.Run("with descend sorting on custom attribute", func(t *testing.T) {
		car1 := NewMongoItem(CarCustomDescend{Brand: "Toyota", Category: "SUV", Seating: []Seater{{Material: "Leather", Occupancy: 2}, {Material: "Leather", Occupancy: 3}, {Material: "Leather", Occupancy: 2}}})
		car2 := NewMongoItem(CarCustomDescend{Brand: "Honda", Category: "Sedan", Seating: []Seater{{Material: "Fabric", Occupancy: 2}, {Material: "Fabric", Occupancy: 3}}})
		car3 := NewMongoItem(CarCustomDescend{Brand: "Ford", Category: "Truck", Seating: []Seater{{Material: "Vinyl", Occupancy: 2}, {Material: "Vinyl", Occupancy: 2}}})
		car4 := NewMongoItem(CarCustomDescend{Brand: "BMW", Category: "Coupe", Seating: []Seater{{Material: "Leather", Occupancy: 2}, {Material: "Leather", Occupancy: 2}}})
		car5 := NewMongoItem(CarCustomDescend{Brand: "Tesla", Category: "Electric", Seating: []Seater{{Material: "Vegan Leather", Occupancy: 2}, {Material: "Vegan Leather", Occupancy: 3}}})

		carRandIds := []string{
			car1.GetRandId(),
			car2.GetRandId(),
			car3.GetRandId(),
			car4.GetRandId(),
			car5.GetRandId(),
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZRevRange(key+descendingTrailing+"ranking", 0, -1).SetVal(carRandIds)

		itemCacheMock := mock_interfaces.NewMockItemCache[CarCustomDescend](ctrl)
		itemCacheMock.EXPECT().Get(car1.GetRandId()).Return(car1, nil)
		itemCacheMock.EXPECT().Get(car2.GetRandId()).Return(car2, nil)
		itemCacheMock.EXPECT().Get(car3.GetRandId()).Return(car3, nil)
		itemCacheMock.EXPECT().Get(car4.GetRandId()).Return(car4, nil)
		itemCacheMock.EXPECT().Get(car5.GetRandId()).Return(car5, nil)

		pagination := initTestPaginationType[CarCustomDescend](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)
		assert.NotNil(t, pagination.sorting)
		assert.Equal(t, "ranking", pagination.sorting.attribute)
		assert.Equal(t, descending, pagination.sorting.direction)

		fetchAll, errorFetchAll := pagination.FetchAll(
			paginationParameters,
			func(item CarCustomDescend, items *[]CarCustomDescend) {

				fmt.Println(item)

				*items = append(*items, item)
			})
		assert.Nil(t, errorFetchAll)
		assert.NotNil(t, fetchAll)
		assert.Equal(t, 5, len(fetchAll))
	})
	t.Run("with ascend sorting on default attribute", func(t *testing.T) {
		car1 := NewMongoItem(CarDefaultAscend{Brand: "Toyota", Category: "SUV", Seating: []Seater{{Material: "Leather", Occupancy: 2}, {Material: "Leather", Occupancy: 3}, {Material: "Leather", Occupancy: 2}}})
		car2 := NewMongoItem(CarDefaultAscend{Brand: "Honda", Category: "Sedan", Seating: []Seater{{Material: "Fabric", Occupancy: 2}, {Material: "Fabric", Occupancy: 3}}})
		car3 := NewMongoItem(CarDefaultAscend{Brand: "Ford", Category: "Truck", Seating: []Seater{{Material: "Vinyl", Occupancy: 2}, {Material: "Vinyl", Occupancy: 2}}})
		car4 := NewMongoItem(CarDefaultAscend{Brand: "BMW", Category: "Coupe", Seating: []Seater{{Material: "Leather", Occupancy: 2}, {Material: "Leather", Occupancy: 2}}})
		car5 := NewMongoItem(CarDefaultAscend{Brand: "Tesla", Category: "Electric", Seating: []Seater{{Material: "Vegan Leather", Occupancy: 2}, {Material: "Vegan Leather", Occupancy: 3}}})

		carRandIds := []string{
			car1.GetRandId(),
			car2.GetRandId(),
			car3.GetRandId(),
			car4.GetRandId(),
			car5.GetRandId(),
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZRange(key+ascendingTrailing+"createdat", 0, -1).SetVal(carRandIds)

		itemCacheMock := mock_interfaces.NewMockItemCache[CarDefaultAscend](ctrl)
		itemCacheMock.EXPECT().Get(car1.GetRandId()).Return(car1, nil)
		itemCacheMock.EXPECT().Get(car2.GetRandId()).Return(car2, nil)
		itemCacheMock.EXPECT().Get(car3.GetRandId()).Return(car3, nil)
		itemCacheMock.EXPECT().Get(car4.GetRandId()).Return(car4, nil)
		itemCacheMock.EXPECT().Get(car5.GetRandId()).Return(car5, nil)

		pagination := initTestPaginationType[CarDefaultAscend](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)
		assert.NotNil(t, pagination.sorting)
		assert.Equal(t, "createdat", pagination.sorting.attribute)
		assert.Equal(t, ascending, pagination.sorting.direction)

		fetchAll, errorFetchAll := pagination.FetchAll(
			paginationParameters,
			func(item CarDefaultAscend, items *[]CarDefaultAscend) {

				fmt.Println(item)

				*items = append(*items, item)
			})
		assert.Nil(t, errorFetchAll)
		assert.NotNil(t, fetchAll)
		assert.Equal(t, 5, len(fetchAll))
	})
	t.Run("with descend sorting on default attribute", func(t *testing.T) {
		car1 := NewMongoItem(Car{Brand: "Toyota", Category: "SUV", Seating: []Seater{{Material: "Leather", Occupancy: 2}, {Material: "Leather", Occupancy: 3}, {Material: "Leather", Occupancy: 2}}})
		car2 := NewMongoItem(Car{Brand: "Honda", Category: "Sedan", Seating: []Seater{{Material: "Fabric", Occupancy: 2}, {Material: "Fabric", Occupancy: 3}}})
		car3 := NewMongoItem(Car{Brand: "Ford", Category: "Truck", Seating: []Seater{{Material: "Vinyl", Occupancy: 2}, {Material: "Vinyl", Occupancy: 2}}})
		car4 := NewMongoItem(Car{Brand: "BMW", Category: "Coupe", Seating: []Seater{{Material: "Leather", Occupancy: 2}, {Material: "Leather", Occupancy: 2}}})
		car5 := NewMongoItem(Car{Brand: "Tesla", Category: "Electric", Seating: []Seater{{Material: "Vegan Leather", Occupancy: 2}, {Material: "Vegan Leather", Occupancy: 3}}})

		carRandIds := []string{
			car1.GetRandId(),
			car2.GetRandId(),
			car3.GetRandId(),
			car4.GetRandId(),
			car5.GetRandId(),
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZRevRange(key+descendingTrailing+"createdat", 0, -1).SetVal(carRandIds)

		itemCacheMock := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCacheMock.EXPECT().Get(car1.GetRandId()).Return(car1, nil)
		itemCacheMock.EXPECT().Get(car2.GetRandId()).Return(car2, nil)
		itemCacheMock.EXPECT().Get(car3.GetRandId()).Return(car3, nil)
		itemCacheMock.EXPECT().Get(car4.GetRandId()).Return(car4, nil)
		itemCacheMock.EXPECT().Get(car5.GetRandId()).Return(car5, nil)

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)
		assert.Nil(t, pagination.sorting)

		fetchAll, errorFetchAll := pagination.FetchAll(
			paginationParameters,
			func(item Car, items *[]Car) {

				fmt.Println(item)

				*items = append(*items, item)
			})
		assert.Nil(t, errorFetchAll)
		assert.NotNil(t, fetchAll)
		assert.Equal(t, 5, len(fetchAll))
	})
	t.Run("with ascend sorting on custom attribute", func(t *testing.T) {
		car1 := NewMongoItem(CarCustomAscend{Brand: "Toyota", Category: "SUV", Seating: []Seater{{Material: "Leather", Occupancy: 2}, {Material: "Leather", Occupancy: 3}, {Material: "Leather", Occupancy: 2}}})
		car2 := NewMongoItem(CarCustomAscend{Brand: "Honda", Category: "Sedan", Seating: []Seater{{Material: "Fabric", Occupancy: 2}, {Material: "Fabric", Occupancy: 3}}})
		car3 := NewMongoItem(CarCustomAscend{Brand: "Ford", Category: "Truck", Seating: []Seater{{Material: "Vinyl", Occupancy: 2}, {Material: "Vinyl", Occupancy: 2}}})
		car4 := NewMongoItem(CarCustomAscend{Brand: "BMW", Category: "Coupe", Seating: []Seater{{Material: "Leather", Occupancy: 2}, {Material: "Leather", Occupancy: 2}}})
		car5 := NewMongoItem(CarCustomAscend{Brand: "Tesla", Category: "Electric", Seating: []Seater{{Material: "Vegan Leather", Occupancy: 2}, {Material: "Vegan Leather", Occupancy: 3}}})

		carRandIds := []string{
			car1.GetRandId(),
			car2.GetRandId(),
			car3.GetRandId(),
			car4.GetRandId(),
			car5.GetRandId(),
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZRange(key+ascendingTrailing+"ranking", 0, -1).SetVal(carRandIds)

		itemCacheMock := mock_interfaces.NewMockItemCache[CarCustomAscend](ctrl)
		itemCacheMock.EXPECT().Get(car1.GetRandId()).Return(car1, nil)
		itemCacheMock.EXPECT().Get(car2.GetRandId()).Return(car2, nil)
		itemCacheMock.EXPECT().Get(car3.GetRandId()).Return(car3, nil)
		itemCacheMock.EXPECT().Get(car4.GetRandId()).Return(car4, nil)
		itemCacheMock.EXPECT().Get(car5.GetRandId()).Return(car5, nil)

		pagination := initTestPaginationType[CarCustomAscend](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)
		assert.NotNil(t, pagination.sorting)
		assert.Equal(t, "ranking", pagination.sorting.attribute)
		assert.Equal(t, ascending, pagination.sorting.direction)

		fetchAll, errorFetchAll := pagination.FetchAll(
			paginationParameters,
			func(item CarCustomAscend, items *[]CarCustomAscend) {

				fmt.Println(item)

				*items = append(*items, item)
			})
		assert.Nil(t, errorFetchAll)
		assert.NotNil(t, fetchAll)
		assert.Equal(t, 5, len(fetchAll))
	})
	t.Run("zrevrange error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZRevRange(key+descendingTrailing+"createdat", 0, -1).SetErr(errors.New("fatal error: Redis connection lost"))

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			nil,
		)

		fetchAll, errorFetchAll := pagination.FetchAll(paginationParameters, nil)
		assert.NotNil(t, errorFetchAll)
		assert.Nil(t, fetchAll)
		assert.Equal(t, REDIS_FATAL_ERROR, errorFetchAll.Err)
	})
	t.Run("Get item redis fatal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZRevRange(key+descendingTrailing+"createdat", 0, -1).SetVal(carRandIds)

		itemCacheMock := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCacheMock.EXPECT().Get(car1.GetRandId()).Return(Car{}, &types.PaginationError{Err: REDIS_FATAL_ERROR})

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)

		fetchAll, errorFetchAll := pagination.FetchAll(
			paginationParameters,
			func(item Car, items *[]Car) {

				fmt.Println(item)

				*items = append(*items, item)
			})

		assert.NotNil(t, errorFetchAll)
		assert.Nil(t, fetchAll)
		assert.Equal(t, REDIS_FATAL_ERROR, errorFetchAll.Err)
	})
	t.Run("One of the item member keys doesn't exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZRevRange(key+descendingTrailing+"createdat", 0, -1).SetVal(carRandIds)

		itemCacheMock := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCacheMock.EXPECT().Get(car1.GetRandId()).Return(Car{}, &types.PaginationError{Err: KEY_NOT_FOUND})
		itemCacheMock.EXPECT().Get(car2.GetRandId()).Return(car2, nil)
		itemCacheMock.EXPECT().Get(car3.GetRandId()).Return(car3, nil)
		itemCacheMock.EXPECT().Get(car4.GetRandId()).Return(car4, nil)
		itemCacheMock.EXPECT().Get(car5.GetRandId()).Return(car5, nil)

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)

		fetchAll, errorFetchAll := pagination.FetchAll(
			paginationParameters,
			func(item Car, items *[]Car) {

				fmt.Println(item)

				*items = append(*items, item)
			})

		assert.Nil(t, errorFetchAll)
		assert.NotNil(t, fetchAll)
		assert.Equal(t, 4, len(fetchAll))
	})
}

func TestFetchLinked(t *testing.T) {
	t.Run("successfully fetch first page without processing", func(t *testing.T) {
		itemPerPage := int64(5)
		car1 := NewMongoItem(Car{Brand: "Toyota", Category: "SUV", Seating: []Seater{{Material: "Leather", Occupancy: 2}, {Material: "Leather", Occupancy: 3}, {Material: "Leather", Occupancy: 2}}})
		car2 := NewMongoItem(Car{Brand: "Honda", Category: "Sedan", Seating: []Seater{{Material: "Fabric", Occupancy: 2}, {Material: "Fabric", Occupancy: 3}}})
		car3 := NewMongoItem(Car{Brand: "Ford", Category: "Truck", Seating: []Seater{{Material: "Vinyl", Occupancy: 2}, {Material: "Vinyl", Occupancy: 2}}})
		car4 := NewMongoItem(Car{Brand: "BMW", Category: "Coupe", Seating: []Seater{{Material: "Leather", Occupancy: 2}, {Material: "Leather", Occupancy: 2}}})
		car5 := NewMongoItem(Car{Brand: "Tesla", Category: "Electric", Seating: []Seater{{Material: "Vegan Leather", Occupancy: 2}, {Material: "Vegan Leather", Occupancy: 3}}})

		carRandIds := []string{
			car5.GetRandId(),
			car4.GetRandId(),
			car3.GetRandId(),
			car2.GetRandId(),
			car1.GetRandId(),
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		itemCacheMock := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCacheMock.EXPECT().Get(car1.GetRandId()).Return(car1, nil)
		itemCacheMock.EXPECT().Get(car2.GetRandId()).Return(car2, nil)
		itemCacheMock.EXPECT().Get(car3.GetRandId()).Return(car3, nil)
		itemCacheMock.EXPECT().Get(car4.GetRandId()).Return(car4, nil)
		itemCacheMock.EXPECT().Get(car5.GetRandId()).Return(car5, nil)

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZRevRange(key+descendingTrailing+"createdat", 0, itemPerPage-1).SetVal(carRandIds)
		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)

		cars, errorFetch := pagination.FetchLinked(
			paginationParameters,
			nil,
			itemPerPage,
			nil,
		)
		assert.Nil(t, errorFetch)
		assert.Equal(t, 5, len(cars))
		assert.Equal(t, car5.GetRandId(), cars[0].GetRandId())
		assert.Equal(t, car4.GetRandId(), cars[1].GetRandId())
		assert.Equal(t, car3.GetRandId(), cars[2].GetRandId())
		assert.Equal(t, car2.GetRandId(), cars[3].GetRandId())
		assert.Equal(t, car1.GetRandId(), cars[4].GetRandId())
	})
	t.Run("successfully fetch second page with processing", func(t *testing.T) {
		itemPerPage := int64(5)
		car1 := NewMongoItem(Car{Brand: "Toyota", Category: "SUV", Seating: []Seater{{Material: "Leather", Occupancy: 2}, {Material: "Leather", Occupancy: 3}, {Material: "Leather", Occupancy: 2}}})
		car2 := NewMongoItem(Car{Brand: "Honda", Category: "Sedan", Seating: []Seater{{Material: "Fabric", Occupancy: 2}, {Material: "Fabric", Occupancy: 3}}})
		car3 := NewMongoItem(Car{Brand: "Ford", Category: "Truck", Seating: []Seater{{Material: "Vinyl", Occupancy: 2}, {Material: "Vinyl", Occupancy: 2}}})
		car4 := NewMongoItem(Car{Brand: "BMW", Category: "Coupe", Seating: []Seater{{Material: "Leather", Occupancy: 2}, {Material: "Leather", Occupancy: 2}}})
		car5 := NewMongoItem(Car{Brand: "Tesla", Category: "Electric", Seating: []Seater{{Material: "Vegan Leather", Occupancy: 2}, {Material: "Vegan Leather", Occupancy: 3}}})

		carRandIds := []string{
			car5.GetRandId(),
			car4.GetRandId(),
			car3.GetRandId(),
			car2.GetRandId(),
			car1.GetRandId(),
		}

		references := []string{
			RandId(),
			RandId(),
			RandId(),
			RandId(),
			RandId(),
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		itemCacheMock := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCacheMock.EXPECT().Get(car1.GetRandId()).Return(car1, nil)
		itemCacheMock.EXPECT().Get(car2.GetRandId()).Return(car2, nil)
		itemCacheMock.EXPECT().Get(car3.GetRandId()).Return(car3, nil)
		itemCacheMock.EXPECT().Get(car4.GetRandId()).Return(car4, nil)
		itemCacheMock.EXPECT().Get(car5.GetRandId()).Return(car5, nil)

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZRevRank(key+descendingTrailing+"createdat", references[4]).SetVal(int64(4))
		mockedRedis.ExpectZRevRange(key+descendingTrailing+"createdat", 5, 5+itemPerPage-1).SetVal(carRandIds)

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)

		cars, errorFetch := pagination.FetchLinked(
			paginationParameters,
			references,
			itemPerPage,
			func(item Car, items *[]Car) {

				item.Category = "SUV"

				*items = append(*items, item)
			},
		)
		assert.Nil(t, errorFetch)
		assert.Equal(t, 5, len(cars))
		assert.Equal(t, "SUV", cars[0].Category)
		assert.Equal(t, "SUV", cars[1].Category)
		assert.Equal(t, "SUV", cars[2].Category)
		assert.Equal(t, "SUV", cars[3].Category)
		assert.Equal(t, "SUV", cars[4].Category)
	})
	t.Run("eleminate one item from result set with processor", func(t *testing.T) {
		itemPerPage := int64(5)
		car1 := NewMongoItem(Car{Brand: "Toyota", Category: "SUV", Seating: []Seater{{Material: "Leather", Occupancy: 2}, {Material: "Leather", Occupancy: 3}, {Material: "Leather", Occupancy: 2}}})
		car2 := NewMongoItem(Car{Brand: "Honda", Category: "Sedan", Seating: []Seater{{Material: "Fabric", Occupancy: 2}, {Material: "Fabric", Occupancy: 3}}})
		car3 := NewMongoItem(Car{Brand: "Ford", Category: "Truck", Seating: []Seater{{Material: "Vinyl", Occupancy: 2}, {Material: "Vinyl", Occupancy: 2}}})
		car4 := NewMongoItem(Car{Brand: "BMW", Category: "Coupe", Seating: []Seater{{Material: "Leather", Occupancy: 2}, {Material: "Leather", Occupancy: 2}}})
		car5 := NewMongoItem(Car{Brand: "Tesla", Category: "Electric", Seating: []Seater{{Material: "Vegan Leather", Occupancy: 2}, {Material: "Vegan Leather", Occupancy: 3}}})

		carRandIds := []string{
			car5.GetRandId(),
			car4.GetRandId(),
			car3.GetRandId(),
			car2.GetRandId(),
			car1.GetRandId(),
		}

		references := []string{
			RandId(),
			RandId(),
			RandId(),
			RandId(),
			RandId(),
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		itemCacheMock := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCacheMock.EXPECT().Get(car1.GetRandId()).Return(car1, nil)
		itemCacheMock.EXPECT().Get(car2.GetRandId()).Return(car2, nil)
		itemCacheMock.EXPECT().Get(car3.GetRandId()).Return(car3, nil)
		itemCacheMock.EXPECT().Get(car4.GetRandId()).Return(car4, nil)
		itemCacheMock.EXPECT().Get(car5.GetRandId()).Return(car5, nil)

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZRevRank(key+descendingTrailing+"createdat", references[4]).SetVal(int64(4))
		mockedRedis.ExpectZRevRange(key+descendingTrailing+"createdat", 5, 5+itemPerPage-1).SetVal(carRandIds)

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)

		cars, errorFetch := pagination.FetchLinked(
			paginationParameters,
			references,
			itemPerPage,
			func(item Car, items *[]Car) {
				// will eleminate SUV from result set
				if item.Category != "SUV" {
					*items = append(*items, item)
				}
			},
		)
		assert.Nil(t, errorFetch)
		assert.Equal(t, 4, len(cars))
		for _, item := range cars {
			assert.NotEqual(t, "SUV", item.Category)
		}
	})
	t.Run("with descend sorting on custom attribute", func(t *testing.T) {
		itemPerPage := int64(5)
		car1 := NewMongoItem(CarCustomDescend{Brand: "Toyota", Ranking: 1, Category: "SUV", Seating: []Seater{{Material: "Leather", Occupancy: 2}, {Material: "Leather", Occupancy: 3}, {Material: "Leather", Occupancy: 2}}})
		car2 := NewMongoItem(CarCustomDescend{Brand: "Honda", Ranking: 2, Category: "Sedan", Seating: []Seater{{Material: "Fabric", Occupancy: 2}, {Material: "Fabric", Occupancy: 3}}})
		car3 := NewMongoItem(CarCustomDescend{Brand: "Ford", Ranking: 3, Category: "Truck", Seating: []Seater{{Material: "Vinyl", Occupancy: 2}, {Material: "Vinyl", Occupancy: 2}}})
		car4 := NewMongoItem(CarCustomDescend{Brand: "BMW", Ranking: 4, Category: "Coupe", Seating: []Seater{{Material: "Leather", Occupancy: 2}, {Material: "Leather", Occupancy: 2}}})
		car5 := NewMongoItem(CarCustomDescend{Brand: "Tesla", Ranking: 5, Category: "Electric", Seating: []Seater{{Material: "Vegan Leather", Occupancy: 2}, {Material: "Vegan Leather", Occupancy: 3}}})

		carRandIds := []string{
			car5.GetRandId(),
			car4.GetRandId(),
			car3.GetRandId(),
			car2.GetRandId(),
			car1.GetRandId(),
		}

		references := []string{
			RandId(),
			RandId(),
			RandId(),
			RandId(),
			RandId(),
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		itemCacheMock := mock_interfaces.NewMockItemCache[CarCustomDescend](ctrl)
		itemCacheMock.EXPECT().Get(car1.GetRandId()).Return(car1, nil)
		itemCacheMock.EXPECT().Get(car2.GetRandId()).Return(car2, nil)
		itemCacheMock.EXPECT().Get(car3.GetRandId()).Return(car3, nil)
		itemCacheMock.EXPECT().Get(car4.GetRandId()).Return(car4, nil)
		itemCacheMock.EXPECT().Get(car5.GetRandId()).Return(car5, nil)

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZRevRank(key+descendingTrailing+"ranking", references[4]).SetVal(int64(4))
		mockedRedis.ExpectZRevRange(key+descendingTrailing+"ranking", 5, 5+itemPerPage-1).SetVal(carRandIds)

		pagination := initTestPaginationType[CarCustomDescend](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)
		cars, errorFetch := pagination.FetchLinked(
			paginationParameters,
			references,
			itemPerPage,
			nil,
		)
		assert.Nil(t, errorFetch)
		assert.Equal(t, 5, len(cars))
		assert.Equal(t, car5.GetRandId(), cars[0].GetRandId())
		assert.Equal(t, car4.GetRandId(), cars[1].GetRandId())
		assert.Equal(t, car3.GetRandId(), cars[2].GetRandId())
		assert.Equal(t, car2.GetRandId(), cars[3].GetRandId())
		assert.Equal(t, car1.GetRandId(), cars[4].GetRandId())
	})
	t.Run("with ascend sorting on default attribute", func(t *testing.T) {
		itemPerPage := int64(5)
		car1 := NewMongoItem(CarDefaultAscend{Brand: "Toyota", Ranking: 6, Category: "SUV", Seating: []Seater{{Material: "Leather", Occupancy: 2}, {Material: "Leather", Occupancy: 3}, {Material: "Leather", Occupancy: 2}}})
		car2 := NewMongoItem(CarDefaultAscend{Brand: "Honda", Ranking: 7, Category: "Sedan", Seating: []Seater{{Material: "Fabric", Occupancy: 2}, {Material: "Fabric", Occupancy: 3}}})
		car3 := NewMongoItem(CarDefaultAscend{Brand: "Ford", Ranking: 8, Category: "Truck", Seating: []Seater{{Material: "Vinyl", Occupancy: 2}, {Material: "Vinyl", Occupancy: 2}}})
		car4 := NewMongoItem(CarDefaultAscend{Brand: "BMW", Ranking: 9, Category: "Coupe", Seating: []Seater{{Material: "Leather", Occupancy: 2}, {Material: "Leather", Occupancy: 2}}})
		car5 := NewMongoItem(CarDefaultAscend{Brand: "Tesla", Ranking: 10, Category: "Electric", Seating: []Seater{{Material: "Vegan Leather", Occupancy: 2}, {Material: "Vegan Leather", Occupancy: 3}}})

		carRandIds := []string{
			car1.GetRandId(),
			car2.GetRandId(),
			car3.GetRandId(),
			car4.GetRandId(),
			car5.GetRandId(),
		}

		references := []string{
			RandId(),
			RandId(),
			RandId(),
			RandId(),
			RandId(),
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		itemCacheMock := mock_interfaces.NewMockItemCache[CarDefaultAscend](ctrl)
		itemCacheMock.EXPECT().Get(car1.GetRandId()).Return(car1, nil)
		itemCacheMock.EXPECT().Get(car2.GetRandId()).Return(car2, nil)
		itemCacheMock.EXPECT().Get(car3.GetRandId()).Return(car3, nil)
		itemCacheMock.EXPECT().Get(car4.GetRandId()).Return(car4, nil)
		itemCacheMock.EXPECT().Get(car5.GetRandId()).Return(car5, nil)

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZRank(key+ascendingTrailing+"createdat", references[4]).SetVal(int64(4))
		mockedRedis.ExpectZRange(key+ascendingTrailing+"createdat", 5, 5+itemPerPage-1).SetVal(carRandIds)

		pagination := initTestPaginationType[CarDefaultAscend](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)
		cars, errorFetch := pagination.FetchLinked(
			paginationParameters,
			references,
			itemPerPage,
			nil,
		)
		assert.Nil(t, errorFetch)
		assert.Equal(t, 5, len(cars))
		assert.Equal(t, car1.GetRandId(), cars[0].GetRandId())
		assert.Equal(t, car2.GetRandId(), cars[1].GetRandId())
		assert.Equal(t, car3.GetRandId(), cars[2].GetRandId())
		assert.Equal(t, car4.GetRandId(), cars[3].GetRandId())
		assert.Equal(t, car5.GetRandId(), cars[4].GetRandId())
	})
	t.Run("with ascend sorting on custom attribute", func(t *testing.T) {
		itemPerPage := int64(5)
		car1 := NewMongoItem(CarCustomAscend{Brand: "Toyota", Ranking: 6, Category: "SUV", Seating: []Seater{{Material: "Leather", Occupancy: 2}, {Material: "Leather", Occupancy: 3}, {Material: "Leather", Occupancy: 2}}})
		car2 := NewMongoItem(CarCustomAscend{Brand: "Honda", Ranking: 7, Category: "Sedan", Seating: []Seater{{Material: "Fabric", Occupancy: 2}, {Material: "Fabric", Occupancy: 3}}})
		car3 := NewMongoItem(CarCustomAscend{Brand: "Ford", Ranking: 8, Category: "Truck", Seating: []Seater{{Material: "Vinyl", Occupancy: 2}, {Material: "Vinyl", Occupancy: 2}}})
		car4 := NewMongoItem(CarCustomAscend{Brand: "BMW", Ranking: 9, Category: "Coupe", Seating: []Seater{{Material: "Leather", Occupancy: 2}, {Material: "Leather", Occupancy: 2}}})
		car5 := NewMongoItem(CarCustomAscend{Brand: "Tesla", Ranking: 10, Category: "Electric", Seating: []Seater{{Material: "Vegan Leather", Occupancy: 2}, {Material: "Vegan Leather", Occupancy: 3}}})

		carRandIds := []string{
			car1.GetRandId(),
			car2.GetRandId(),
			car3.GetRandId(),
			car4.GetRandId(),
			car5.GetRandId(),
		}

		references := []string{
			RandId(),
			RandId(),
			RandId(),
			RandId(),
			RandId(),
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		itemCacheMock := mock_interfaces.NewMockItemCache[CarCustomAscend](ctrl)
		itemCacheMock.EXPECT().Get(car1.GetRandId()).Return(car1, nil)
		itemCacheMock.EXPECT().Get(car2.GetRandId()).Return(car2, nil)
		itemCacheMock.EXPECT().Get(car3.GetRandId()).Return(car3, nil)
		itemCacheMock.EXPECT().Get(car4.GetRandId()).Return(car4, nil)
		itemCacheMock.EXPECT().Get(car5.GetRandId()).Return(car5, nil)

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZRank(key+ascendingTrailing+"ranking", references[4]).SetVal(int64(4))
		mockedRedis.ExpectZRange(key+ascendingTrailing+"ranking", 5, 5+itemPerPage-1).SetVal(carRandIds)

		pagination := initTestPaginationType[CarCustomAscend](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)
		cars, errorFetch := pagination.FetchLinked(
			paginationParameters,
			references,
			itemPerPage,
			nil,
		)
		assert.Nil(t, errorFetch)
		assert.Equal(t, 5, len(cars))
		assert.Equal(t, car1.GetRandId(), cars[0].GetRandId())
		assert.Equal(t, car2.GetRandId(), cars[1].GetRandId())
		assert.Equal(t, car3.GetRandId(), cars[2].GetRandId())
		assert.Equal(t, car4.GetRandId(), cars[3].GetRandId())
		assert.Equal(t, car5.GetRandId(), cars[4].GetRandId())
	})
	t.Run("zrank error", func(t *testing.T) {
		itemPerPage := int64(5)
		references := []string{
			RandId(),
			RandId(),
			RandId(),
			RandId(),
			RandId(),
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		itemCacheMock := mock_interfaces.NewMockItemCache[CarCustomAscend](ctrl)

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZRank(key+ascendingTrailing+"ranking", references[4]).SetErr(errors.New("fatal error: Redis connection lost"))

		pagination := initTestPaginationType[CarCustomAscend](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)
		cars, errorFetch := pagination.FetchLinked(
			paginationParameters,
			references,
			itemPerPage,
			nil,
		)
		assert.NotNil(t, errorFetch)
		assert.Nil(t, cars)
		assert.Equal(t, REDIS_FATAL_ERROR, errorFetch.Err)
	})
	t.Run("zrevrank error", func(t *testing.T) {
		itemPerPage := int64(5)
		references := []string{
			RandId(),
			RandId(),
			RandId(),
			RandId(),
			RandId(),
		}

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		itemCacheMock := mock_interfaces.NewMockItemCache[CarCustomDescend](ctrl)

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZRank(key+descendingTrailing+"ranking", references[4]).SetErr(errors.New("fatal error: Redis connection lost"))

		pagination := initTestPaginationType[CarCustomDescend](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)
		cars, errorFetch := pagination.FetchLinked(
			paginationParameters,
			references,
			itemPerPage,
			nil,
		)
		assert.NotNil(t, errorFetch)
		assert.Nil(t, cars)
		assert.Equal(t, REDIS_FATAL_ERROR, errorFetch.Err)
	})
}
