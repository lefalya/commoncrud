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
	t.Run("remove item success with no database specified", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redis, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key).SetVal(20)
		mockedRedis.ExpectZRem(key, car.GetRandId()).SetVal(1)

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
	t.Run("zcard fatal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key).SetErr(errors.New("fatal error: Redis connection lost"))

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			nil,
		)

		errorRemoveItem := pagination.RemoveItem(paginationParameters, car)
		assert.NotNil(t, errorRemoveItem)
		assert.Equal(t, REDIS_FATAL_ERROR, errorRemoveItem.Err)
	})
	t.Run("zrem but item not found", func(t *testing.T) {
		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZCard(key).SetVal(3)
		mockedRedis.ExpectZRem(key, car.GetRandId()).SetVal(0)

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			nil,
		)

		errorRemoveItem := pagination.RemoveItem(paginationParameters, car)
		assert.Nil(t, errorRemoveItem)
	})
	t.Run("itemcache delete error", func(t *testing.T) {
	})
	t.Run("mongo delete error", func(t *testing.T) {})
}

func TestTotalItemOnCache(t *testing.T) {

	t.Run("", func(t *testing.T) {})
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
		mockedRedis.ExpectZRevRange(key, 0, -1).SetVal(carRandIds)

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
			func(item Car, items *[]Car, args ...interface{}) {

				fmt.Println(item)

				*items = append(*items, item)
			},
			nil)

		assert.Nil(t, errorFetchAll)
		assert.NotNil(t, fetchAll)
		assert.Equal(t, 5, len(fetchAll))
	})

	t.Run("zrevrange error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZRevRange(key, 0, -1).SetErr(errors.New("fatal error: Redis connection lost"))

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			nil,
		)

		fetchAll, errorFetchAll := pagination.FetchAll(
			paginationParameters,
			nil,
			nil)
		assert.NotNil(t, errorFetchAll)
		assert.Nil(t, fetchAll)
		assert.Equal(t, REDIS_FATAL_ERROR, errorFetchAll.Err)
		assert.Equal(t, "fetchall.zrevrange_fatal_error", errorFetchAll.Context)
	})

	t.Run("Get item redis fatal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZRevRange(key, 0, -1).SetVal(carRandIds)

		itemCacheMock := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCacheMock.EXPECT().Get(car1.GetRandId()).Return(Car{}, &commonlogger.CommonError{Err: REDIS_FATAL_ERROR})

		pagination := initTestPaginationType[Car](
			paginationKeyFormat,
			itemKeyFormat,
			logger,
			redisDB,
			itemCacheMock,
		)

		fetchAll, errorFetchAll := pagination.FetchAll(
			paginationParameters,
			func(item Car, items *[]Car, args ...interface{}) {

				fmt.Println(item)

				*items = append(*items, item)
			},
			nil)

		assert.NotNil(t, errorFetchAll)
		assert.Nil(t, fetchAll)
		assert.Equal(t, REDIS_FATAL_ERROR, errorFetchAll.Err)
		assert.Equal(t, "fetchall.get_item_fatal_error", errorFetchAll.Context)
	})

	t.Run("One of the item member keys doesn't exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		redisDB, mockedRedis := redismock.NewClientMock()
		mockedRedis.ExpectZRevRange(key, 0, -1).SetVal(carRandIds)

		itemCacheMock := mock_interfaces.NewMockItemCache[Car](ctrl)
		itemCacheMock.EXPECT().Get(car1.GetRandId()).Return(Car{}, &commonlogger.CommonError{Err: KEY_NOT_FOUND})
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
			func(item Car, items *[]Car, args ...interface{}) {

				fmt.Println(item)

				*items = append(*items, item)
			},
			nil)

		assert.Nil(t, errorFetchAll)
		assert.NotNil(t, fetchAll)
		assert.Equal(t, 4, len(fetchAll))
	})
}
