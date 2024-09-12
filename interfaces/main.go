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

type ItemCache[T Item] interface {
	Get(randId string) (T, *commonlogger.CommonError)
	Set(item T) *commonlogger.CommonError
	Delete(item T) *commonlogger.CommonError
}

// individual key format: individualKeyFormat:[item.RandId]
// Functions only ask pagination key parameters.
type Pagination[T Item] interface {
	WithMongo(mongo Mongo[T], paginationFilter bson.A)
	ItemPerPage() int64
	AddItem(pagKeyParams []string, item *T) *commonlogger.CommonError
	UpdateItem(item *T) *commonlogger.CommonError
	RemoveItem(pagKeyParams []string, item *T) *commonlogger.CommonError
	TotalItem(keyFormat string, keyParameters []string) *commonlogger.CommonError
	FetchLinked(pagKeyParams []string, references []string, processor PaginationProcessor[T], processorArgs ...interface{}) ([]T, *commonlogger.CommonError)
	FetchAll(pagKeyParams []string, processor PaginationProcessor[T], processorArgs ...interface{}) ([]T, *commonlogger.CommonError)
	SeedPartial(
		paginationKeyParameters []string,
		lastItem T,
		paginationFilter bson.A,
		processor PaginationProcessor[T],
		processorArgs ...interface{}) ([]T, *commonlogger.CommonError)
	SeedAll() ([]T, *commonlogger.CommonError)
}

type PaginationProcessor[T Item] func(item T, items *[]T, args ...interface{})

type Mongo[T Item] interface {
	Create(item T) *commonlogger.CommonError
	FindOne(randId string) (T, *commonlogger.CommonError)
	FindMany(filter bson.D, findOptions *options.FindOptions) (*mongo.Cursor, *commonlogger.CommonError)
	Update(item T) *commonlogger.CommonError
	Delete(item T) *commonlogger.CommonError
	SetPaginationFilter(filter bson.A)
	GetPaginationFilter() bson.A
}
