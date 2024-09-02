package interfaces

import (
	"github.com/lefalya/commonlogger"
)

type Processor[T any] func(value *T, args ...string)

type LinkedPagination[T any] interface {
	AddItem(
		keyFormat string,
		score float64,
		value *T,
		contextPrefix string,
		parameters ...interface{},
	) *commonlogger.CommonError

	RemoveItem(
		keyFormat string,
		member string,
		value *T,
		contextPrefix string,
		parameters ...interface{},
	) *commonlogger.CommonError

	TotalItem(
		keyFormat string,
		contextPrefix string,
		parameters ...interface{},
	) *commonlogger.CommonError

	Paginate(
		keyFormat string,
		individualKeyFormat string,
		contextPrefix string,
		lastMember []string,
		filter Processor[T],
		parameters ...interface{},
	) (
		[]T,
		string,
		int64,
		*commonlogger.CommonError)
}

type NumericPagination[T any] interface{}

type MongoSeeder[T any] interface {
	SeedLinkedPagination(
		keyFormat string,
		individualKeyFormat string,
		subtraction int64,
		lastMember string,
		filter Processor[T],
		parameters ...interface{},
	) (
		[]T,
		*commonlogger.CommonError,
	)
}
