package commoncrud

import (
	"context"
	"log/slog"
	"reflect"

	"github.com/lefalya/commoncrud/interfaces"
	"github.com/lefalya/commoncrud/types"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ascending  = "ascending"
	descending = "descending"
	randomized = iota

	ascendingTrailing  = ":asc"
	descendingTrailing = ":desc"
)

type SortingOption struct {
	attribute string
	direction string
	index     int
}

type PaginationType[T interfaces.Item] struct {
	pagKeyFormat  string
	itemKeyFormat string
	logger        *slog.Logger
	redisClient   redis.UniversalClient
	sorting       *SortingOption
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

	sortingOpt := SortingOption{}
	var isSortingDirExists bool

	// Get the type of T using reflection
	t := reflect.TypeOf((*T)(nil)).Elem()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		if f.Tag.Get("bson") != "" {
			sortingOpt.attribute = f.Tag.Get("bson")
		} else if f.Tag.Get("db") != "" {
			sortingOpt.attribute = f.Tag.Get("db")
		}
		sorting := f.Tag.Get("sorting")
		if sorting != "" {
			isSortingDirExists = true
			sortingOpt.index = i
			if sorting == ascending {
				sortingOpt.direction = ascending
			} else if sorting == descending {
				sortingOpt.direction = descending
			}
		}

	}

	pagination := &PaginationType[T]{
		pagKeyFormat:  pagKeyFormat,
		itemKeyFormat: itemKeyFormat,
		logger:        logger,
		redisClient:   redisClient,
		itemCache:     itemCache,
	}

	if isSortingDirExists {
		pagination.sorting = &sortingOpt
	}

	return pagination
}

func (pg *PaginationType[T]) SortBy(attribute string, direction string) {
	sortingOpt := SortingOption{
		attribute: attribute,
		direction: direction,
	}

	pg.sorting = &sortingOpt
}

func (pg *PaginationType[T]) WithMongo(mongo interfaces.Mongo[T], paginationFilter bson.A) {
	pg.mongo = mongo
	pg.mongo.SetPaginationFilter(paginationFilter)
}

func (pg *PaginationType[T]) AddItem(pagKeyParams []string, item T) *types.PaginationError {
	if pg.mongo != nil {
		err := pg.mongo.Create(item)
		if err != nil {
			return &types.PaginationError{
				Err:     err.Err,
				Details: err.Details,
				Message: "Failed to create item to MongoDB",
			}
		}
	}

	key := concatKey(pg.pagKeyFormat, pagKeyParams)
	var sortedSetKey string
	if pg.sorting != nil && pg.sorting.direction == ascending {
		sortedSetKey = key + ":" + ascending
	} else {
		sortedSetKey = key + ":" + descending
	}

	totalItem := pg.redisClient.ZCard(
		context.TODO(),
		sortedSetKey,
	)
	if totalItem.Err() != nil {
		return &types.PaginationError{
			Err:     REDIS_FATAL_ERROR,
			Details: totalItem.Err().Error(),
			Message: "Failed to count total items on Redis",
		}
	}

	// only add item to sorted set, if the sorted set exists
	if totalItem.Val() > 0 {
		errorSet := pg.itemCache.Set(item)
		if errorSet != nil {
			return &types.PaginationError{
				Err:     errorSet.Err,
				Details: errorSet.Details,
				Message: "Failed to set item to Redis",
			}
		}

		var score float64
		if pg.sorting != nil {
			value, ok := reflect.ValueOf(item).Elem().Field(pg.sorting.index).Interface().(float64)
			if ok {
				score = value
			} else {
				score = float64(item.GetCreatedAt().Unix())
			}
		} else {
			score = float64(item.GetCreatedAt().Unix())
		}

		sortedSetMember := redis.Z{
			Score:  score,
			Member: item.GetRandId(),
		}
		setSortedSet := pg.redisClient.ZAdd(
			context.TODO(),
			sortedSetKey,
			sortedSetMember,
		)
		if setSortedSet.Err() != nil {
			return &types.PaginationError{
				Err:     REDIS_FATAL_ERROR,
				Details: setSortedSet.Err().Error(),
				Message: "Failed to add item to pagination set on Redis",
			}
		}

		setExpire := pg.redisClient.Expire(
			context.TODO(),
			sortedSetKey,
			SORTED_SET_TTL,
		)
		if setExpire.Err() != nil {
			return &types.PaginationError{
				Err:     REDIS_FATAL_ERROR,
				Details: setExpire.Err().Error(),
				Message: "Failed to extend pagination set expiration on Redis",
			}
		}
	}

	return nil
}

func (pg *PaginationType[T]) UpdateItem(item T) *types.PaginationError {
	if pg.mongo != nil {
		err := pg.mongo.Update(item)
		if err != nil {
			return &types.PaginationError{
				Err:     err.Err,
				Details: err.Details,
				Message: "Failed to update item on MongoDB",
			}
		}
	}

	errorSet := pg.itemCache.Set(item)
	if errorSet != nil {
		return errorSet
	}

	return nil
}

func (pg *PaginationType[T]) RemoveItem(pagKeyParams []string, item T) *types.PaginationError {
	key := concatKey(pg.pagKeyFormat, pagKeyParams)

	if pg.mongo != nil {
		err := pg.mongo.Delete(item)
		if err != nil {
			return &types.PaginationError{
				Err:     err.Err,
				Details: err.Details,
				Message: "Failed to delete item from MongoDB",
			}
		}
	}

	errorDelete := pg.itemCache.Del(item)
	if errorDelete != nil {
		return errorDelete
	}

	var sortedSetKey string
	if pg.sorting != nil && pg.sorting.direction == ascending {
		sortedSetKey = key + ":" + ascending
	} else {
		sortedSetKey = key + ":" + descending
	}

	totalItem := pg.redisClient.ZCard(
		context.TODO(),
		sortedSetKey,
	)
	if totalItem.Err() != nil {
		return &types.PaginationError{
			Err:     REDIS_FATAL_ERROR,
			Details: totalItem.Err().Error(),
			Message: "Failed to count total items on Redis",
		}
	}

	// only remove item from sorted set, if the sorted set exists
	if totalItem.Val() > 0 {
		removeItemSortedSet := pg.redisClient.ZRem(context.TODO(), sortedSetKey, item.GetRandId())

		if removeItemSortedSet.Err() != nil {
			return &types.PaginationError{
				Err:     REDIS_FATAL_ERROR,
				Details: removeItemSortedSet.Err().Error(),
				Message: "Failed to remove item from pagination set on Redis",
			}
		}
	}

	// will not return an error if the ZRem, Del, or Delete command results in zero deletions.
	return nil
}

func (pg *PaginationType[T]) TotalItemOnCache(pagKeyParams []string) *types.PaginationError {
	key := concatKey(pg.pagKeyFormat, pagKeyParams)

	var sortedSetKey string
	if pg.sorting != nil && pg.sorting.direction == ascending {
		sortedSetKey = key + ":" + ascending
	} else {
		sortedSetKey = key + ":" + descending
	}

	totalItem := pg.redisClient.ZCard(
		context.TODO(),
		sortedSetKey,
	)

	if totalItem.Err() != nil {
		return &types.PaginationError{
			Err:     REDIS_FATAL_ERROR,
			Details: totalItem.Err().Error(),
			Message: "Failed to count total items on Redis",
		}
	}

	return nil
}

func (pg *PaginationType[T]) FetchOne(randId string) (*T, *types.PaginationError) {
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
	processor interfaces.PaginationProcessor[T]) ([]T, *types.PaginationError) {
	var items []T
	var start int64
	var stop int64

	key := concatKey(pg.pagKeyFormat, pagKeyParams)
	totalReferences := len(references)

	if totalReferences > 0 {
		if totalReferences > MAXIMUM_AMOUNT_REFERENCES {
			return nil, &types.PaginationError{
				Err:     TOO_MUCH_REFERENCES,
				Message: "Too much references!",
			}
		}

		for i := len(references) - 1; i >= 0; i-- {
			var rank *redis.IntCmd
			if pg.sorting != nil && pg.sorting.direction == ascending {
				sortedSetKey := key + ":" + ascending
				rank = pg.redisClient.ZRank(context.TODO(), sortedSetKey, references[i])
			} else {
				sortedSetKey := key + ":" + descending
				rank = pg.redisClient.ZRevRank(context.TODO(), sortedSetKey, references[i])
			}

			if rank.Err() != nil {
				if rank.Err() == redis.Nil {
					continue
				}

				return nil, &types.PaginationError{
					Err:     REDIS_FATAL_ERROR,
					Details: rank.Err().Error(),
					Message: "Failed to get reference's index from pagination set on Redis",
				}
			}

			start = rank.Val() + 1
			break
		}

		if start == 0 {
			return nil, &types.PaginationError{
				Err:     NO_VALID_REFERENCES,
				Message: "No references found from pagination set on Redis",
			}
		}
	}

	stop = start + itemPerPage - 1

	var members *redis.StringSliceCmd
	if pg.sorting != nil && pg.sorting.direction == ascending {
		members = pg.redisClient.ZRange(context.TODO(), key, start, stop)
	} else {
		members = pg.redisClient.ZRevRange(context.TODO(), key, start, stop)
	}

	if members.Err() != nil {
		return nil, &types.PaginationError{
			Err:     REDIS_FATAL_ERROR,
			Details: members.Err().Error(),
			Message: "Failed to get items from pagination set on Redis",
		}
	}

	if len(members.Val()) > 0 {
		pg.redisClient.Expire(context.TODO(), key, SORTED_SET_TTL)

		for _, member := range members.Val() {
			item, errorGetItem := pg.itemCache.Get(member)
			if errorGetItem != nil && errorGetItem.Err == REDIS_FATAL_ERROR {
				return nil, &types.PaginationError{
					Err:     errorGetItem.Err,
					Details: errorGetItem.Details,
					Message: "Failed to get item details from Redis",
				}
			} else if errorGetItem != nil && errorGetItem.Err == KEY_NOT_FOUND {
				continue
			}

			if processor != nil {
				processor(item, &items)
			} else {
				items = append(items, item)
			}
		}
	}

	return items, nil
}

func (pg *PaginationType[T]) FetchAll(pagKeyParams []string, processor interfaces.PaginationProcessor[T]) ([]T, *types.PaginationError) {
	var items []T
	key := concatKey(pg.pagKeyFormat, pagKeyParams)

	var members *redis.StringSliceCmd
	if pg.sorting != nil && pg.sorting.direction == ascending {
		sortedSetKey := key + ":" + ascending
		members = pg.redisClient.ZRange(context.TODO(), sortedSetKey, 0, -1)
	} else {
		sortedSetKey := key + ":" + descending
		members = pg.redisClient.ZRevRange(context.TODO(), sortedSetKey, 0, -1)
	}

	if members.Err() != nil {
		return nil, &types.PaginationError{
			Err:     REDIS_FATAL_ERROR,
			Details: members.Err().Error(),
			Message: "Failed to get items from pagination set on Redis",
		}
	}

	if len(members.Val()) > 0 {
		pg.redisClient.Expire(context.TODO(), key, SORTED_SET_TTL)

		for _, member := range members.Val() {
			item, errorGetItem := pg.itemCache.Get(member)
			if errorGetItem != nil && errorGetItem.Err == REDIS_FATAL_ERROR {
				return nil, &types.PaginationError{
					Err:     errorGetItem.Err,
					Details: errorGetItem.Details,
					Message: "Failed to get item details from Redis",
				}
			} else if errorGetItem != nil && errorGetItem.Err == KEY_NOT_FOUND {
				continue
			}

			if processor != nil {
				processor(item, &items)
			} else {
				items = append(items, item)
			}
		}
	}

	return items, nil
}

func (pg *PaginationType[T]) SeedOne(randId string) (*T, *types.PaginationError) {
	var result T
	if pg.mongo != nil {
		item, errorFind := pg.mongo.FindOne(randId)
		if errorFind != nil {
			if errorFind.Err == MONGO_FATAL_ERROR {
				return nil, &types.PaginationError{
					Err:     errorFind.Err,
					Details: errorFind.Details,
					Message: "Item not found on MongoDB",
				}
			} else {
				// MONGO_FATAL_ERROR
				return nil, &types.PaginationError{
					Err:     errorFind.Err,
					Details: errorFind.Details,
					Message: "Fatal error from MongoDB while finding item",
				}
			}
		}
		result = item
	} else {
		return nil, &types.PaginationError{
			Err:     NO_DATABASE_CONFIGURED,
			Message: "No database configured",
		}
	}

	return &result, nil
}

func (pg *PaginationType[T]) SeedLinked(
	paginationKeyParameters []string,
	reference T,
	itemPerPage int64,
	processor interfaces.SeedProcessor[T],
) ([]T, *types.PaginationError) {
	var result []T
	var filter bson.D

	if pg.mongo != nil {
		if !reflect.ValueOf(reference).IsZero() {
			inMongoItem, ok := any(reference).(interfaces.MongoItem)
			if !ok {
				// return lastItem must be in MongoItem
				return nil, &types.PaginationError{
					Err:     LASTITEM_MUST_MONGOITEM,
					Message: "Using MongoDB as database but the reference is not in MongoItem",
				}
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
		if pg.sorting != nil && pg.sorting.direction == ascending {
			findOptions.SetSort(bson.D{{"_id", 1}})
		} else {
			findOptions.SetSort(bson.D{{"_id", -1}})
		}

		findOptions.SetLimit(itemPerPage)

		items, errorFindItems := pg.mongo.FindMany(
			filter,
			findOptions,
			pg,
			paginationKeyParameters,
			processor,
		)

		if errorFindItems != nil {
			return nil, &types.PaginationError{
				Err:     errorFindItems.Err,
				Details: errorFindItems.Details,
				Message: "MongoDB fatal error while retrieving items",
			}
		}

		result = items
	} else {
		return nil, &types.PaginationError{
			Err:     NO_DATABASE_CONFIGURED,
			Message: "No database configured",
		}
	}

	return result, nil
}

func (pg *PaginationType[T]) SeedAll(
	paginationKeyParameters []string,
	processor interfaces.SeedProcessor[T],
) ([]T, *types.PaginationError) {
	var results []T
	var filter bson.D

	if pg.mongo != nil {
		filter = bson.D{
			{Key: "$and", Value: pg.mongo.GetPaginationFilter()},
		}

		findOptions := options.Find()
		if pg.sorting != nil && pg.sorting.direction == ascending {
			findOptions.SetSort(bson.D{{"_id", 1}})
		} else {
			findOptions.SetSort(bson.D{{"_id", -1}})
		}

		cursor, errorFindItems := pg.mongo.FindMany(
			filter,
			findOptions,
			pg,
			paginationKeyParameters,
			processor,
		)

		if errorFindItems != nil {
			return nil, &types.PaginationError{
				Err:     errorFindItems.Err,
				Details: errorFindItems.Details,
				Message: "MongoDB fatal error while retrieving items",
			}
		}

		results = cursor
	} else {
		return nil, &types.PaginationError{
			Err:     NO_DATABASE_CONFIGURED,
			Message: "No database configured",
		}
	}

	return results, nil
}
