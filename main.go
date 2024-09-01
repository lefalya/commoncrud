package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/lefalya/commonlogger"
	"github.com/lefalya/commonpagination/interfaces"
	commonRedisInterfaces "github.com/lefalya/commonredis/interfaces"
	"github.com/redis/go-redis/v9"

	loggerInterfaces "github.com/lefalya/commonlogger/interfaces"
	loggerSchema "github.com/lefalya/commonlogger/schema"
)

const (
	DAY_TTL            = 24 * time.Hour
	INDIVIDUAL_KEY_TTL = DAY_TTL * 7
	SORTED_SET_TTL     = DAY_TTL * 2

	MAX_LAST_MEMBERS = 5
	ITEM_PER_PAGE    = 45
)

var (
	TOO_MUCH_RANDIDS = errors.New("10000; ")
)

func joinParameters(parameters ...interface{}) string {

	stringSlice := make([]string, len(parameters))
	for i, v := range parameters {
		stringSlice[i] = fmt.Sprint(v)
	}

	return strings.Join(stringSlice, ", ")
}

type LinkedPagination[T any] struct {
	redisClient *redis.Client
	logger      *slog.Logger
	generic     commonRedisInterfaces.Generic[T]
}

func NewLinkedPagination[T any](
	logger *slog.Logger,
	redisClient *redis.Client,
	generic commonRedisInterfaces.Generic[T],
) *LinkedPagination[T] {

	return &LinkedPagination[T]{
		logger:      logger,
		redisClient: redisClient,
		generic:     generic,
	}
}

func (cr *LinkedPagination[T]) AddItem(
	keyFormat string,
	score float64,
	member string,
	value *T,
	contextPrefix string,
	errorHandler loggerInterfaces.LogHelper,
	parameters ...interface{},
) *loggerSchema.CommonError {

	finalKey := fmt.Sprintf(keyFormat, parameters...)

	sortedSetMember := redis.Z{
		Score:  score,
		Member: member,
	}

	setSortedSet := cr.redisClient.ZAdd(
		context.TODO(),
		finalKey,
		sortedSetMember,
	)

	if setSortedSet.Err() != nil {

		return errorHandler(
			cr.logger,
			FATAL_ERROR,
			setSortedSet.Err().Error(),
			contextPrefix+".set_sorted_set_fatal_error",
			*value,
		)
	}

	setExpire := cr.redisClient.Expire(
		context.TODO(),
		finalKey,
		SORTED_SET_TTL,
	)

	if setExpire.Err() != nil {

		return errorHandler(
			cr.logger,
			FATAL_ERROR,
			setExpire.Err().Error(),
			contextPrefix+".set_sorted_set_expire_fatal_error",
			*value,
		)
	}

	return nil
}

func (cr *LinkedPagination[T]) RemoveItem(
	keyFormat string,
	member string,
	value *T,
	contextPrefix string,
	errorHandler loggerInterfaces.LogHelper,
	parameters ...interface{},
) *loggerSchema.CommonError {

	finalKey := fmt.Sprintf(keyFormat, parameters...)

	removeFromSortedSet := cr.redisClient.ZRem(
		context.TODO(),
		finalKey,
		member,
	)

	if removeFromSortedSet.Err() != nil {

		return errorHandler(
			cr.logger,
			FATAL_ERROR,
			removeFromSortedSet.Err().Error(),
			contextPrefix+".delete_from_sorted_set_fatal_error",
			*value,
		)
	}

	return nil
}

func (cr *LinkedPagination[T]) TotalItem(
	keyFormat string,
	contextPrefix string,
	parameters ...interface{},
) *loggerSchema.CommonError {

	finalKey := fmt.Sprintf(keyFormat, parameters...)

	totalItemsOnSortedSet := cr.redisClient.ZCard(
		context.TODO(),
		finalKey,
	)

	if totalItemsOnSortedSet.Err() != nil {

		return commonlogger.LogError(
			cr.logger,
			FATAL_ERROR,
			totalItemsOnSortedSet.Err().Error(),
			contextPrefix+".get_total_item_sorted_set_fatal_error",
			"keyFormat", keyFormat,
			"contextPrefix", contextPrefix,
			"finalKey", finalKey,
		)
	}

	return nil
}

func (cr *LinkedPagination[T]) Paginate(
	keyFormat string,
	contextPrefix string,
	lastMembers []string,
	individualKeyFormat string,
	processor interfaces.Processor[T],
	processorArgs []string,
	parameters ...interface{},
) (
	[]T,
	string,
	int64,
	*loggerSchema.CommonError) {

	finalKey := fmt.Sprintf(keyFormat, parameters...)

	var returnValues []T
	var validLastMembers string
	var start int64
	var stop int64

	totalRandIds := len(lastMembers)

	if totalRandIds > 0 {

		if totalRandIds > MAX_LAST_MEMBERS {

			commonError := commonlogger.LogError(
				cr.logger,
				TOO_MUCH_RANDIDS,
				"Too much randIds inserted",
				contextPrefix+".MAX_LAST_MEMBERS_EXCEEDED",
				"keyFormat", keyFormat,
				"finalKey", finalKey,
				"len(lastMembers)", strconv.Itoa(len(lastMembers)),
				"parameters", joinParameters(parameters),
			)

			return nil, validLastMembers, start, commonError
		}

		for i := len(lastMembers) - 1; i >= 0; i-- {

			rank := cr.redisClient.ZRevRank(
				context.TODO(),
				finalKey,
				lastMembers[i],
			)

			if rank.Err() != nil {

				// if the collection is deleted
				if rank.Err() == redis.Nil {

					continue
				}

				commonError := commonlogger.LogError(
					cr.logger,
					FATAL_ERROR,
					rank.Err().Error(),
					contextPrefix+".ZREVRANK_FATAL_ERROR",
					"keyFormat", keyFormat,
					"finalKey", finalKey,
					"len(lastMembers)", strconv.Itoa(len(lastMembers)),
					"parameters", joinParameters(parameters),
				)

				return nil, validLastMembers, start, commonError
			}

			validLastMembers = lastMembers[i]
			start = rank.Val() + 1
			break
		}

	}

	stop = start + ITEM_PER_PAGE - 1
	members := cr.redisClient.ZRevRange(
		context.TODO(),
		finalKey,
		start,
		stop,
	)

	if members.Err() != nil {

		commonError := commonlogger.LogError(
			cr.logger,
			FATAL_ERROR,
			members.Err().Error(),
			contextPrefix+".ZREVRANGE_FATAL_ERROR",
			"keyFormat", keyFormat,
			"finalKey", finalKey,
			"parameters", joinParameters(parameters),
		)

		return nil, validLastMembers, start, commonError
	}

	if len(members.Val()) > 0 {

		cr.redisClient.Expire(
			context.TODO(),
			finalKey,
			SORTED_SET_TTL,
		)

		for _, member := range members.Val() {

			memberAsT, errorGetItem := cr.generic.Get(
				individualKeyFormat,
				contextPrefix+".paginate.",
				member,
			)

			if errorGetItem != nil {

				return nil, validLastMembers, start, errorGetItem
			}

			processor(memberAsT, processorArgs...)

			validLastMembers = member
			returnValues = append(returnValues, *memberAsT)
		}
	}

	return returnValues, validLastMembers, start, nil
}
