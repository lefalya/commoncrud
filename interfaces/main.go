package interfaces

import (
	"time"

	"github.com/lefalya/commonlogger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

// individual key format: individualKeyFormat:[item.RandId]
// Functions only ask pagination key parameters.
type Pagination[T Item] interface {
	WithMongo(mongo Mongo[T], paginationFilter bson.A)
	AddItem(pagKeyParams []string, item T) *commonlogger.CommonError
	UpdateItem(item T) *commonlogger.CommonError
	RemoveItem(pagKeyParams []string, item T) *commonlogger.CommonError
	TotalItemOnCache(pagKeyParams []string) *commonlogger.CommonError
	FetchOne(randId string) (*T, *commonlogger.CommonError)
	FetchLinked(
		pagKeyParams []string,
		references []string,
		itemPerPage int64,
		processor PaginationProcessor[T],
		processorArgs ...interface{}) ([]T, *commonlogger.CommonError)
	FetchAll(pagKeyParams []string, processor PaginationProcessor[T], processorArgs ...interface{}) ([]T, *commonlogger.CommonError)
	SeedOne(randId string) (*T, *commonlogger.CommonError)
	SeedPartial(
		paginationKeyParameters []string,
		lastItem T,
		itemPerPage int64,
		processor PaginationProcessor[T],
		processorArgs ...interface{}) ([]T, *commonlogger.CommonError)
	SeedAll(
		paginationKeyParameters []string,
		processor PaginationProcessor[T],
		processorArgs ...interface{}) ([]T, *commonlogger.CommonError)
}

type PaginationProcessor[T Item] func(item T, items *[]T, args ...interface{})

type ItemCache[T Item] interface {
	Get(randId string) (T, *commonlogger.CommonError)
	Set(item T) *commonlogger.CommonError
	Del(item T) *commonlogger.CommonError
}

type Mongo[T Item] interface {
	Create(item T) *commonlogger.CommonError
	FindOne(randId string) (T, *commonlogger.CommonError)
	FindMany(filter bson.D, findOptions *options.FindOptions) (*mongo.Cursor, *commonlogger.CommonError)
	Update(item T) *commonlogger.CommonError
	Delete(item T) *commonlogger.CommonError
	SetPaginationFilter(filter bson.A)
	GetPaginationFilter() bson.A
}
