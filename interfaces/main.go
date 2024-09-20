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
	AddItem(pagKeyParams []string, item T) *types.PaginationError
	UpdateItem(item T) *types.PaginationError
	RemoveItem(pagKeyParams []string, item T) *types.PaginationError
	TotalItemOnCache(pagKeyParams []string) *types.PaginationError
	FetchOne(randId string) (*T, *types.PaginationError)
	FetchLinked(
		pagKeyParams []string,
		references []string,
		itemPerPage int64,
		processor PaginationProcessor[T],
	) ([]T, *types.PaginationError)
	FetchAll(pagKeyParams []string, processor PaginationProcessor[T]) ([]T, *types.PaginationError)
	SeedOne(randId string) (*T, *types.PaginationError)
	SeedLinked(
		paginationKeyParameters []string,
		lastItem T,
		itemPerPage int64,
		processor SeedProcessor[T],
	) ([]T, *types.PaginationError)
	SeedAll(paginationKeyParameters []string, processor SeedProcessor[T]) ([]T, *types.PaginationError)
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
	Update(item T) *types.PaginationError
	Delete(item T) *types.PaginationError
	SetPaginationFilter(filter bson.A)
	GetPaginationFilter() bson.A
}
