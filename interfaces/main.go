package interfaces

import (
	"time"

	"github.com/lefalya/commoncrud/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	AddItem(item T, paginationParameters ...string) *types.PaginationError
	UpdateItem(item T, paginationParameters ...string) *types.PaginationError
	RemoveItem(item T, paginationParameters ...string) *types.PaginationError
	TotalItemOnCache(paginationParameters ...string) *types.PaginationError
	FetchOne(randId string) (*T, *types.PaginationError)
	FetchLinked(
		references []string,
		processor PaginationProcessor[T],
		paginationParameters ...string,
	) ([]T, *types.PaginationError)
	FetchAll(processor PaginationProcessor[T], paginationParameters ...string) ([]T, *types.PaginationError)
	SeedOne(randId string) (*T, *types.PaginationError)
	SeedLinked(
		lastItem T,
		processor SeedProcessor[T],
		paginationParameters ...string,
	) ([]T, *types.PaginationError)
	SeedAll(processor SeedProcessor[T], paginationParameters ...string) ([]T, *types.PaginationError)
	SeedCardinality(paginationParameters ...string) *types.PaginationError
}

type PaginationProcessor[T Item] func(item T, items *[]T)
type SeedProcessor[T Item] func(item *T)

type ItemCache[T Item] interface {
	Get(randId string) (T, *types.PaginationError)
	Set(item T) *types.PaginationError
	Del(item T) *types.PaginationError
}

type Mongo[T Item] interface {
	Create(item T) *types.PaginationError
	FindOne(randId string) (T, *types.PaginationError)
	FindMany(
		filter bson.D,
		findOptions *options.FindOptions,
		pagination Pagination[T],
		pagKeyParams []string,
		seedProcessor SeedProcessor[T],
	) ([]T, *types.PaginationError)
	Count(filter bson.D, pagination Pagination[T], paginationParameters []string) (int64, *types.PaginationError)
	Update(item T) *types.PaginationError
	Delete(item T) *types.PaginationError
	SetPaginationFilter(filter bson.A)
	GetPaginationFilter() bson.A
}
