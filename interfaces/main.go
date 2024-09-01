package main

import (
	loggerInterfaces "github.com/lefalya/commonlogger/interfaces"
	loggerSchema "github.com/lefalya/commonlogger/schema"
)

type LinkedPagination[T any] interface {
	AddItem(
		keyFormat string,
		score float64,
		value *T,
		contextPrefix string,
		errorHandler *loggerInterfaces.LogHelper,
		parameters ...interface{},
	) *loggerSchema.CommonError
	RemoveItem(
		keyFormat string,
		member string,
		value *T,
		contextPrefix string,
		errorHandler *loggerInterfaces.LogHelper,
		parameters ...interface{},
	) *loggerSchema.CommonError
	TotalItem(
		keyFormat string,
		contextPrefix string,
		parameters ...interface{},
	) *loggerSchema.CommonError
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
		*loggerSchema.CommonError)
}
