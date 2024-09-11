package interfaces

import (
	"time"

	"github.com/lefalya/commonlogger"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Item interface {
	SetUUID()
	GetUUID() string
	SetRandId()
	GetRandId() string
	SetCreatedAt(time time.Time)
	GetCreatedAt() time.Time
	SetUpdatedAt(time time.Time)
	GetUpdatedAt() time.Time
	SetCreatedAtString(timeString string)
	GetCreatedAtString() string
	SetUpdatedAtString(timeString string)
	GetUpdatedAtString() string
}

type MongoItem interface {
	Item
	SetObjectId()
	GetObjectId() primitive.ObjectID
}

type ItemCache[T Item] interface {
	Get(randId string) (T, *commonlogger.CommonError)
	Set(item T) *commonlogger.CommonError
	Delete(item T) *commonlogger.CommonError
}

// individual key format: individualKeyFormat:[item.RandId]
// Functions only ask pagination key parameters.
type Pagination[T Item] interface {
	ItemPerPage() int64
	AddItem(pagKeyParams []string, item *T) *commonlogger.CommonError
	RemoveItem(pagKeyParams []string, item *T) *commonlogger.CommonError
	TotalItem(keyFormat string, keyParameters []string) *commonlogger.CommonError
	FetchLinked(pagKeyParams []string, references []string, processor PaginationProcessor[T], processorArgs ...interface{}) ([]T, *commonlogger.CommonError)
	FetchAll(pagKeyParams []string, processor PaginationProcessor[T], processorArgs ...interface{}) ([]T, *commonlogger.CommonError)
}

type PaginationProcessor[T Item] func(item T, items []T, redisClient *redis.Client, args ...interface{})

type Mongo[T Item] interface {
	Create(item T) *commonlogger.CommonError
	Find(randId string) (T, *commonlogger.CommonError)
	Update(item T) *commonlogger.CommonError
	Delete(item T) *commonlogger.CommonError
	SeedLinked(
		paginationKeyParameters []string,
		subtraction int64,
		referenceAsHex string,
		lastItem T,
		paginationFilter bson.A,
		processor MongoProcesor[T],
		processorArgs ...interface{}) ([]T, *commonlogger.CommonError)
}

type MongoProcesor[T Item] func(item *T, redisCLient *redis.Client, args ...interface{})
