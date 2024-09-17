package commoncrud

import (
	"context"
	"log/slog"
	"reflect"
	"strings"
	"time"

	"github.com/lefalya/commoncrud/interfaces"
	"github.com/lefalya/commonlogger"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type PaginationType[T interfaces.Item] struct {
	pagKeyFormat  string
	itemKeyFormat string
	logger        *slog.Logger
	redisClient   redis.UniversalClient
	mongo         interfaces.Mongo[T]
	itemCache     interfaces.ItemCache[T]
}

func Pagination[T interfaces.Item](
	pagKeyFormat string,
	itemKeyFormat string,
	logger *slog.Logger,
	redisClient redis.UniversalClient,
) *PaginationType[T] {
	itemCache := &ItemCacheType[T]{
		logger:        logger,
		redisClient:   redisClient,
		itemKeyFormat: itemKeyFormat,
	}

	return &PaginationType[T]{
		pagKeyFormat:  pagKeyFormat,
		itemKeyFormat: itemKeyFormat,
		logger:        logger,
		redisClient:   redisClient,
		itemCache:     itemCache,
	}
}

func (pg *PaginationType[T]) WithMongo(mongo interfaces.Mongo[T], paginationFilter bson.A) {
	pg.mongo = mongo
	pg.mongo.SetPaginationFilter(paginationFilter)
}

func (pg *PaginationType[T]) AddItem(pagKeyParams []string, item T) *commonlogger.CommonError {
	if pg.mongo != nil {
		err := pg.mongo.Create(item)
		if err != nil {
			return err
		}
	}

	key := concatKey(pg.pagKeyFormat, pagKeyParams)

	totalItem := pg.redisClient.ZCard(
		context.TODO(),
		key,
	)
	if totalItem.Err() != nil {
		return ItemLogHelper(
			pg.logger,
			REDIS_FATAL_ERROR,
			totalItem.Err().Error(),
			"additem.zcard_redis_fatal_error",
			item,
			"pagKeyParams", strings.Join(pagKeyParams, ", "),
		)
	}

	// only add item to sorted set, if the sorted set exists
	if totalItem.Val() > 0 {
		errorSet := pg.itemCache.Set(item)
		if errorSet != nil {
			return errorSet
		}

		sortedSetMember := redis.Z{
			Score:  float64(item.GetCreatedAt().Unix()),
			Member: item.GetRandId(),
		}

		setSortedSet := pg.redisClient.ZAdd(
			context.TODO(),
			key,
			sortedSetMember,
		)

		if setSortedSet.Err() != nil {
			return ItemLogHelper(
				pg.logger,
				REDIS_FATAL_ERROR,
				setSortedSet.Err().Error(),
				"additem.zadd_redis_fatal_error",
				item,
				"pagKeyParams", strings.Join(pagKeyParams, ", "),
			)
		}

		setExpire := pg.redisClient.Expire(
			context.TODO(),
			key,
			SORTED_SET_TTL,
		)

		if setExpire.Err() != nil {
			return ItemLogHelper(
				pg.logger,
				REDIS_FATAL_ERROR,
				setExpire.Err().Error(),
				"additem.setexpire_redis_fatal_error",
				item,
				"pagKeyParams", strings.Join(pagKeyParams, ", "),
			)
		}
	}

	return nil
}

func (pg *PaginationType[T]) UpdateItem(item T) *commonlogger.CommonError {
	if pg.mongo != nil {
		err := pg.mongo.Update(item)
		if err != nil {
			return err
		}
	}

	errorSet := pg.itemCache.Set(item)
	if errorSet != nil {
		return errorSet
	}

	return nil
}

func (pg *PaginationType[T]) RemoveItem(pagKeyParams []string, item T) *commonlogger.CommonError {

	key := concatKey(pg.pagKeyFormat, pagKeyParams)

	totalItem := pg.redisClient.ZCard(
		context.TODO(),
		key,
	)
	if totalItem.Err() != nil {
		return ItemLogHelper(
			pg.logger,
			REDIS_FATAL_ERROR,
			totalItem.Err().Error(),
			"removeitem.zcard_redis_fatal_error",
			item,
			"pagKeyParams", strings.Join(pagKeyParams, ", "),
		)
	}

	// only remove item from sorted set, if the sorted set exists
	if totalItem.Val() > 0 {
		removeItemSortedSet := pg.redisClient.ZRem(context.TODO(), key, item.GetRandId())

		if removeItemSortedSet.Err() != nil {
			return ItemLogHelper(
				pg.logger,
				REDIS_FATAL_ERROR,
				removeItemSortedSet.Err().Error(),
				"removeitem.zrem_redis_fatal_error",
				item,
				"pagKeyFormat", pg.pagKeyFormat,
				"pagKeyParams", strings.Join(pagKeyParams, ","),
			)
		}
	}

	errorDelete := pg.itemCache.Del(item)
	if errorDelete != nil {
		return errorDelete
	}

	if pg.mongo != nil {
		err := pg.mongo.Delete(item)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pg *PaginationType[T]) TotalItemOnCache(pagKeyParams []string) *commonlogger.CommonError {
	key := concatKey(pg.pagKeyFormat, pagKeyParams)

	totalItem := pg.redisClient.ZCard(
		context.TODO(),
		key,
	)

	if totalItem.Err() != nil {
		return commonlogger.LogError(
			pg.logger,
			REDIS_FATAL_ERROR,
			totalItem.Err().Error(),
			"totalitem.zcard_redis_fatal_error",
			"pagKeyFormat", pg.pagKeyFormat,
			"pagKeyParams", strings.Join(pagKeyParams, ","),
		)
	}

	return nil
}

func (pg *PaginationType[T]) FetchOne(randId string) (*T, *commonlogger.CommonError) {
	item, errorGet := pg.itemCache.Get(randId)

	if errorGet != nil {
		return nil, errorGet
	}

	return &item, nil
}

func (pg *PaginationType[T]) FetchLinked(
	pagKeyParams []string,
	references []string,
	itemPerPage int64,
	processor interfaces.PaginationProcessor[T],
	processorArgs ...interface{}) ([]T, *commonlogger.CommonError) {
	var items []T
	var start int64
	var stop int64

	errorArgs := []string{"pagKeyParams", strings.Join(pagKeyParams, ","), "references", strings.Join(references, ",")}

	key := concatKey(pg.pagKeyFormat, pagKeyParams)
	totalReferences := len(references)

	if totalReferences > 0 {
		if totalReferences > MAXIMUM_AMOUNT_REFERENCES {
			return nil, commonlogger.LogError(
				pg.logger,
				TOO_MUCH_REFERENCES,
				"too much references!",
				"fetchlinked.too_much_references",
				errorArgs...,
			)
		}

		for i := len(references) - 1; i >= 0; i-- {
			rank := pg.redisClient.ZRevRank(context.TODO(), key, references[i])

			if rank.Err() != nil {
				if rank.Err() == redis.Nil {
					continue
				}

				return nil, commonlogger.LogError(
					pg.logger,
					REDIS_FATAL_ERROR,
					rank.Err().Error(),
					"fetchlinked.zrevrank_fatal_error",
					errorArgs...,
				)
			}

			start = rank.Val() + 1
			break
		}

		if start == 0 {
			return nil, commonlogger.LogError(
				pg.logger,
				NO_VALID_REFERENCES,
				"no valid references found!",
				"fetchlinked.no_valid_references",
				errorArgs...,
			)
		}
	}

	stop = start + itemPerPage - 1

	members := pg.redisClient.ZRevRange(context.TODO(), key, start, stop)
	if members.Err() != nil {
		return nil, commonlogger.LogError(
			pg.logger,
			REDIS_FATAL_ERROR,
			members.Err().Error(),
			"fetchlinked.zrevragne_fatal_error",
			errorArgs...,
		)
	}

	if len(members.Val()) > 0 {
		pg.redisClient.Expire(context.TODO(), key, SORTED_SET_TTL)

		for _, member := range members.Val() {
			item, errorGetItem := pg.itemCache.Get(member)
			if errorGetItem != nil && errorGetItem.Err == REDIS_FATAL_ERROR {
				return nil, commonlogger.LogError(
					pg.logger,
					REDIS_FATAL_ERROR,
					"redis fatal error while retrieving individual key",
					"fetchlinked.get_item_fatal_error",
					errorArgs...,
				)
			} else if errorGetItem != nil && errorGetItem.Err == KEY_NOT_FOUND {
				commonlogger.LogError(
					pg.logger,
					KEY_NOT_FOUND,
					"individual key not found!",
					"fetchlinked.get_item_key_not_found",
					errorArgs...,
				)

				continue
			}

			processor(item, &items, processorArgs...)
		}
	}

	return items, nil
}

func (pg *PaginationType[T]) FetchAll(
	pagKeyParams []string,
	processor interfaces.PaginationProcessor[T],
	processorArgs ...interface{}) ([]T, *commonlogger.CommonError) {
	var items []T
	key := concatKey(pg.pagKeyFormat, pagKeyParams)

	members := pg.redisClient.ZRevRange(context.TODO(), key, 0, -1)
	if members.Err() != nil {
		return nil, commonlogger.LogError(
			pg.logger,
			REDIS_FATAL_ERROR,
			members.Err().Error(),
			"fetchall.zrevrange_fatal_error",
			"pagKeyParams", strings.Join(pagKeyParams, ", "),
		)
	}

	if len(members.Val()) > 0 {
		pg.redisClient.Expire(context.TODO(), key, SORTED_SET_TTL)

		for _, member := range members.Val() {
			item, errorGetItem := pg.itemCache.Get(member)
			if errorGetItem != nil && errorGetItem.Err == REDIS_FATAL_ERROR {
				return nil, commonlogger.LogError(
					pg.logger,
					REDIS_FATAL_ERROR,
					"redis fatal error while retrieving individual key",
					"fetchall.get_item_fatal_error",
					"pagKeyParams", strings.Join(pagKeyParams, ", "),
				)
			} else if errorGetItem != nil && errorGetItem.Err == KEY_NOT_FOUND {
				commonlogger.LogError(
					pg.logger,
					KEY_NOT_FOUND,
					"individual key not found!",
					"fetchall.get_item_key_not_found",
					"pagKeyParams", strings.Join(pagKeyParams, ", "),
				)

				continue
			}

			processor(item, &items, processorArgs...)
		}
	}

	return items, nil
}

func (pg *PaginationType[T]) SeedOne(randId string) (*T, *commonlogger.CommonError) {
	var result T
	if pg.mongo != nil {
		item, errorFind := pg.mongo.FindOne(randId)
		if errorFind != nil {
			return nil, errorFind
		}
		result = item
	} else {
		return nil, commonlogger.LogError(
			pg.logger,
			NO_DATABASE_CONFIGURED,
			"",
			"seedone.no_database_configured",
			"randId", randId,
		)
	}

	return &result, nil
}

func (pg *PaginationType[T]) SeedLinked(
	paginationKeyParameters []string,
	lastItem T,
	itemPerPage int64,
	processor interfaces.PaginationProcessor[T],
	processorArgs ...interface{},
) ([]T, *commonlogger.CommonError) {
	errorArgs := []string{}

	var result []T
	var filter bson.D

	if pg.mongo != nil {
		if !reflect.ValueOf(lastItem).IsZero() {
			inMongoItem, ok := any(lastItem).(interfaces.MongoItem)
			if !ok {
				// return lastItem must be in MongoItem
				return nil, commonlogger.LogError(
					pg.logger,
					LASTITEM_MUST_MONGOITEM,
					"invalid item, not in MongoItem type",
					"seedpartial.lastitem_must_mongoitem",
					errorArgs...,
				)
			}

			filter = bson.D{
				{
					Key: "$and",
					Value: append(
						pg.mongo.GetPaginationFilter(),
						bson.A{
							bson.D{
								{
									Key: "_id",
									Value: bson.D{
										{
											Key:   "$lt",
											Value: inMongoItem.GetObjectId(),
										},
									},
								},
							},
						}...,
					),
				},
			}
		} else {
			if pg.mongo != nil {
				filter = bson.D{
					{Key: "$and", Value: pg.mongo.GetPaginationFilter()},
				}
			}
		}

		findOptions := options.Find()
		findOptions.SetSort(bson.D{{"_id", -1}})
		findOptions.SetLimit(itemPerPage)

		cursor, errorFindItems := pg.mongo.FindMany(
			filter,
			findOptions,
		)

		if errorFindItems != nil {
			return nil, errorFindItems
		}
		defer cursor.Close(context.TODO())

		for cursor.Next(context.TODO()) {
			var item T

			errorDecode := cursor.Decode(&item)
			if errorDecode != nil {
				commonlogger.LogError(
					pg.logger,
					FAILED_DECODE,
					errorDecode.Error(),
					"seedlinkedpagination.failed_decode",
					errorArgs...,
				)

				continue
			}

			parsedTime, errorParse := time.Parse(FORMATTED_TIME, item.GetCreatedAtString())
			if errorParse == nil {
				item.SetCreatedAt(parsedTime)
			}

			parsedTime, errorParse = time.Parse(FORMATTED_TIME, item.GetUpdatedAtString())
			if errorParse == nil {
				item.SetUpdatedAt(parsedTime)
			}

			processor(
				item,
				&result,
				processorArgs...,
			)

			pg.AddItem(paginationKeyParameters, item)

			result = append(result, item)
		}
	} else {
		return nil, commonlogger.LogError(
			pg.logger,
			NO_DATABASE_CONFIGURED,
			"",
			"seedpartial.no_database_configured",
			errorArgs...,
		)
	}

	return result, nil
}

func (pg *PaginationType[T]) SeedAll(
	paginationKeyParameters []string,
	processor interfaces.PaginationProcessor[T],
	processorArgs ...interface{},
) ([]T, *commonlogger.CommonError) {
	errorArgs := []string{}

	var result []T
	var filter bson.D

	if pg.mongo != nil {
		filter = bson.D{
			{Key: "$and", Value: pg.mongo.GetPaginationFilter()},
		}

		findOptions := options.Find()
		findOptions.SetSort(bson.D{{"_id", -1}})

		cursor, errorFindItems := pg.mongo.FindMany(
			filter,
			findOptions,
		)

		if errorFindItems != nil {
			return nil, errorFindItems
		}
		defer cursor.Close(context.TODO())

		for cursor.Next(context.TODO()) {
			var item T

			errorDecode := cursor.Decode(&item)
			if errorDecode != nil {
				commonlogger.LogError(
					pg.logger,
					FAILED_DECODE,
					errorDecode.Error(),
					"seedall.failed_decode",
					errorArgs...,
				)

				continue
			}

			parsedTime, errorParse := time.Parse(FORMATTED_TIME, item.GetCreatedAtString())
			if errorParse == nil {
				item.SetCreatedAt(parsedTime)
			}

			parsedTime, errorParse = time.Parse(FORMATTED_TIME, item.GetUpdatedAtString())
			if errorParse == nil {
				item.SetUpdatedAt(parsedTime)
			}

			processor(
				item,
				&result,
				processorArgs...,
			)

			pg.AddItem(paginationKeyParameters, item)

			result = append(result, item)
		}
	}

	return result, nil
}
