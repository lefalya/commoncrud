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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	DAY_TTL            = 24 * time.Hour
	INDIVIDUAL_KEY_TTL = DAY_TTL * 7
	SORTED_SET_TTL     = DAY_TTL * 2

	MAX_LAST_MEMBERS = 5
	ITEM_PER_PAGE    = 45
)

var (
	REDIS_FATAL_ERROR   = errors.New("Redis fatal error!")
	TOO_MUCH_REFERENCES = errors.New("Too much references!")
	INVALID_HEX         = errors.New("Invalid hex: fail to convert hex to ObjectId")
	REFERENCE_NOT_FOUND = errors.New("Reference not found!")
)

func joinParameters(parameters ...interface{}) string {

	stringSlice := make([]string, len(parameters))
	for i, v := range parameters {
		stringSlice[i] = fmt.Sprint(v)
	}

	return strings.Join(stringSlice, ", ")
}

type LinkedPaginationType[T any] struct {
	redisClient   *redis.Client
	logger        *slog.Logger
	generic       commonRedisInterfaces.Generic[T]
	sortedSetKey  string
	individualKey string
}

func LinkedPagination[T any](
	logger *slog.Logger,
	redisClient *redis.Client,
	generic commonRedisInterfaces.Generic[T],
) *LinkedPaginationType[T] {

	return &LinkedPaginationType[T]{
		logger:      logger,
		redisClient: redisClient,
		generic:     generic,
	}
}

func (cr *LinkedPaginationType[T]) AddItem(
	keyFormat string,
	score float64,
	member string,
	contextPrefix string,
	parameters ...interface{},
) *commonlogger.CommonError {

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
			REDIS_FATAL_ERROR,
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
			REDIS_FATAL_ERROR,
			setExpire.Err().Error(),
			contextPrefix+".set_sorted_set_expire_fatal_error",
			*value,
		)
	}

	return nil
}

func (cr *LinkedPaginationType[T]) RemoveItem(
	keyFormat string,
	member string,
	value *T,
	contextPrefix string,
	parameters ...interface{},
) *commonlogger.CommonError {

	finalKey := fmt.Sprintf(keyFormat, parameters...)

	removeFromSortedSet := cr.redisClient.ZRem(
		context.TODO(),
		finalKey,
		member,
	)

	if removeFromSortedSet.Err() != nil {

		return errorHandler(
			cr.logger,
			REDIS_FATAL_ERROR,
			removeFromSortedSet.Err().Error(),
			contextPrefix+".delete_from_sorted_set_fatal_error",
			*value,
		)
	}

	return nil
}

func (cr *LinkedPaginationType[T]) TotalItem(
	keyFormat string,
	contextPrefix string,
	parameters ...interface{},
) *commonlogger.CommonError {

	finalKey := fmt.Sprintf(keyFormat, parameters...)

	totalItemsOnSortedSet := cr.redisClient.ZCard(
		context.TODO(),
		finalKey,
	)

	if totalItemsOnSortedSet.Err() != nil {

		return commonlogger.LogError(
			cr.logger,
			REDIS_FATAL_ERROR,
			totalItemsOnSortedSet.Err().Error(),
			contextPrefix+".get_total_item_sorted_set_fatal_error",
			"keyFormat", keyFormat,
			"contextPrefix", contextPrefix,
			"finalKey", finalKey,
		)
	}

	return nil
}

func (cr *LinkedPaginationType[T]) Paginate(
	keyFormat string,
	contextPrefix string,
	references []string,
	individualKeyFormat string,
	processor interfaces.Processor[T],
	processorArgs []string,
	parameters ...interface{},
) (
	[]T,
	string,
	int64,
	*commonlogger.CommonError) {

	finalKey := fmt.Sprintf(keyFormat, parameters...)

	var returnValues []T
	var validLastMembers string
	var start int64
	var stop int64

	totalRandIds := len(references)

	if totalRandIds > 0 {

		if totalRandIds > MAX_LAST_MEMBERS {

			commonError := commonlogger.LogError(
				cr.logger,
				TOO_MUCH_REFERENCES,
				"Too much randIds inserted",
				contextPrefix+".MAX_LAST_MEMBERS_EXCEEDED",
				"keyFormat", keyFormat,
				"finalKey", finalKey,
				"len(lastMembers)", strconv.Itoa(len(references)),
				"parameters", joinParameters(parameters),
			)

			return nil, validLastMembers, start, commonError
		}

		for i := len(references) - 1; i >= 0; i-- {

			rank := cr.redisClient.ZRevRank(
				context.TODO(),
				finalKey,
				references[i],
			)

			if rank.Err() != nil {

				// if the collection is deleted
				if rank.Err() == redis.Nil {

					continue
				}

				commonError := commonlogger.LogError(
					cr.logger,
					REDIS_FATAL_ERROR,
					rank.Err().Error(),
					contextPrefix+".ZREVRANK_FATAL_ERROR",
					"keyFormat", keyFormat,
					"finalKey", finalKey,
					"len(lastMembers)", strconv.Itoa(len(references)),
					"parameters", joinParameters(parameters),
				)

				return nil, validLastMembers, start, commonError
			}

			validLastMembers = references[i]
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
			REDIS_FATAL_ERROR,
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

type MongoSeederType[T any] struct {
	redisClient      *redis.Client
	redisGeneric     commonRedisInterfaces.Generic[T]
	mongoCollection  *mongo.Collection
	logger           *slog.Logger
	linkedPagination interfaces.LinkedPagination[T]
}

func MongoSeeder[T any](
	logger *slog.Logger,
	redisClient *redis.Client,
	mongoCollection *mongo.Collection,
	generic commonRedisInterfaces.Generic[T],
	linkedPagination interfaces.LinkedPagination[T],
) *MongoSeederType[T] {

	return &MongoSeederType[T]{
		logger:           logger,
		redisClient:      redisClient,
		mongoCollection:  mongoCollection,
		linkedPagination: linkedPagination,
	}
}

func (ms *MongoSeederType[T]) SeedLinkedPagination(
	individualKeyFormat string,
	subtraction int64,
	referenceAsHex string,
	lastMember string,
	processors interfaces.Processor[T],
	paginationFilter bson.A,
	singularFilter bson.D,
) ([]T, *commonlogger.CommonError) {

	var cursor *mongo.Cursor
	var filter bson.D
	var manyT []T

	findOptions := options.Find()
	findOptions.SetSort(bson.D{{"_id", -1}})

	if subtraction > 0 {

		referenceAsObjectId, errorConvertHex := primitive.ObjectIDFromHex(referenceAsHex)

		if errorConvertHex != nil {

			return nil, commonlogger.LogError(
				ms.logger,
				INVALID_HEX,
				errorConvertHex.Error(),
				"individualKeyFormat", individualKeyFormat,
				"subtraction", strconv.Itoa(int(subtraction)),
				"referenceAsHex", referenceAsHex,
			)
		}

		remainingItem := ITEM_PER_PAGE - subtraction
		findOptions.SetLimit(int64(remainingItem))

		filter = bson.D{
			{Key: "$and",
				Value: append(
					paginationFilter,
					bson.A{
						bson.D{
							{Key: "_id",
								Value: bson.D{
									{
										Key:   "$lt",
										Value: referenceAsObjectId,
									},
								},
							},
						},
					}...,
				),
			},
		}

	} else {

		findOptions.SetLimit(int64(ITEM_PER_PAGE))

		if referenceAsHex != "" {

			// lastItem, _ := finder(lastMember)

			var referenceItem T
			findReferenceError := ms.mongoCollection.FindOne(
				context.TODO(),
				singularFilter,
			).Decode(referenceItem)

			if findReferenceError != nil {

				if findReferenceError == mongo.ErrNoDocuments {

					return nil, REFERENCE_NOT_FOUND
				}

				return nil, findReferenceError
			}

		}
	}
}
