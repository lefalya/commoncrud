package commoncrud

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/golang/mock/gomock"
	"github.com/lefalya/commoncrud/interfaces"
	mock_interfaces "github.com/lefalya/commoncrud/mocks"
	"github.com/lefalya/commonlogger"
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
	return &PaginationType[T]{
		pagKeyFormat:  pagKeyFormat,
		itemKeyFormat: itemKeyFormat,
		logger:        logger,
		redisClient:   redisClient,
		itemCache:     itemCache,
	}
}

type Seater struct {
	Material  string
	Occupancy int64
}

type Car struct {
	*Item
	*MongoItem
	Brand    string
	Category string
	Seating  []Seater
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

func TestConcatKey(t *testing.T) {

}

func TestAddItem(t *testing.T) {
	t.Run("successfully add item without addition to sorted set", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mongoMock := mock_interfaces.NewMockMongo[Car](ctrl)
		mongoMock.EXPECT().SetPaginationFilter(nil)
		mongoMock.EXPECT().Create(car).Return(nil)

		redis, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key).SetVal(0)

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
		mockedRedis.ExpectZCard(key).SetVal(3)
		expectedZMember := redis.Z{
			Score:  float64(car.GetCreatedAt().Unix()),
			Member: car.GetRandId(),
		}
		mockedRedis.ExpectZAdd(key, expectedZMember).SetVal(1)
		mockedRedis.ExpectExpire(key, SORTED_SET_TTL).SetVal(true)

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
	})
	t.Run("successfully add item with no database specified", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		itemCacheMock := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCacheMock.EXPECT().Set(car).Return(nil)

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key).SetVal(3)
		expectedZMember := redis.Z{
			Score:  float64(car.GetCreatedAt().Unix()),
			Member: car.GetRandId(),
		}
		mockedRedis.ExpectZAdd(key, expectedZMember).SetVal(1)
		mockedRedis.ExpectExpire(key, SORTED_SET_TTL).SetVal(true)

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
	t.Run("zcard fatal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mongoMock := mock_interfaces.NewMockMongo[Car](ctrl)
		mongoMock.EXPECT().Create(car).Return(nil)
		mongoMock.EXPECT().SetPaginationFilter(nil)

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key).SetErr(errors.New("fatal error: Redis connection lost"))

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
		itemCacheMock.EXPECT().Set(car).Return(&commonlogger.CommonError{Err: REDIS_FATAL_ERROR})

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key).SetVal(3)

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
		assert.Equal(t, "additem.zcard_redis_fatal_error", errorAddItem.Context)
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
		mockedRedis.ExpectZCard(key).SetVal(3)
		expectedZMember := redis.Z{
			Score:  float64(car.GetCreatedAt().Unix()),
			Member: car.GetRandId(),
		}
		mockedRedis.ExpectZAdd(key, expectedZMember).SetErr(errors.New("fatal error: Redis connection lost"))

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
		assert.Equal(t, "additem.zadd_redis_fatal_error", errorAddItem.Context)
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
		mockedRedis.ExpectZCard(key).SetVal(3)
		expectedZMember := redis.Z{
			Score:  float64(car.GetCreatedAt().Unix()),
			Member: car.GetRandId(),
		}
		mockedRedis.ExpectZAdd(key, expectedZMember).SetErr(errors.New("fatal error: Redis connection lost"))

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
		assert.Equal(t, "additem.zadd_redis_fatal_error", errorAddItem.Context)
	})
	t.Run("mongo create error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mongoMock := mock_interfaces.NewMockMongo[Car](ctrl)
		mongoMock.EXPECT().SetPaginationFilter(nil)
		mongoMock.EXPECT().Create(car).Return(&commonlogger.CommonError{Err: MONGO_FATAL_ERROR})

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
		mongoMock.EXPECT().Update(car).Return(&commonlogger.CommonError{Err: MONGO_FATAL_ERROR})

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
		itemCache.EXPECT().Set(car).Return(&commonlogger.CommonError{Err: REDIS_FATAL_ERROR})

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
		mockedRedis.ExpectZCard(key).SetVal(0)

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

	t.Run("remove item success with no database specified", func(t *testing.T) {})

	t.Run("zcard fatal error", func(t *testing.T) {})

	t.Run("zrem fatal error", func(t *testing.T) {})

	t.Run("itemcache delete error", func(t *testing.T) {})

	t.Run("mongo delete error", func(t *testing.T) {})
}

func TestTotalItemOnCache(t *testing.T) {

	t.Run("", func(t *testing.T) {})
}
