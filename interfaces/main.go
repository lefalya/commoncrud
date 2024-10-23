package interfaces

import (
	"github.com/lefalya/commoncrud/types"
	"time"
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

// individual key format: individualKeyFormat:[item.RandId]
// Functions only ask pagination key parameters.
type Pagination[T Item] interface {
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
