package commoncrud

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/lefalya/commoncrud/interfaces"
	"github.com/lefalya/commoncrud/types"
	"github.com/redis/go-redis/v9"
)

type ItemCacheType[T interfaces.Item] struct {
	itemKeyFormat string
	logger        *slog.Logger
	redisClient   redis.UniversalClient
}

func ItemCache[T interfaces.Item](keyFormat string, logger *slog.Logger, redisClient redis.UniversalClient) *ItemCacheType[T] {
	return &ItemCacheType[T]{
		itemKeyFormat: keyFormat,
		logger:        logger,
		redisClient:   redisClient,
	}
}

func (cr *ItemCacheType[T]) Get(randId string) (T, *types.PaginationError) {
	var nilItem T
	key := fmt.Sprintf(cr.itemKeyFormat, randId)

	result := cr.redisClient.Get(context.TODO(), key)

	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return nilItem, &types.PaginationError{
				Err:     KEY_NOT_FOUND,
				Details: "key not found!",
			}
		}
		return nilItem, &types.PaginationError{
			Err:     REDIS_FATAL_ERROR,
			Details: result.Err().Error(),
		}
	}

	var item T
	errorUnmarshal := json.Unmarshal([]byte(result.Val()), &item)

	parsedTimeCreatedAt, _ := time.Parse(FORMATTED_TIME, item.GetCreatedAtString())
	parsedTimeUpdatedAt, _ := time.Parse(FORMATTED_TIME, item.GetUpdatedAtString())

	item.SetCreatedAt(parsedTimeCreatedAt)
	item.SetUpdatedAt(parsedTimeUpdatedAt)

	if errorUnmarshal != nil {
		return nilItem, &types.PaginationError{
			Err:     ERROR_PARSE_JSON,
			Details: errorUnmarshal.Error(),
		}
	}

	setExpire := cr.redisClient.Expire(context.TODO(), key, INDIVIDUAL_KEY_TTL)
	if setExpire.Err() != nil {
		return nilItem, &types.PaginationError{
			Err:     REDIS_FATAL_ERROR,
			Details: setExpire.Err().Error(),
		}
	}

	return item, nil
}

func (cr *ItemCacheType[T]) Set(item T) *types.PaginationError {
	key := fmt.Sprintf(cr.itemKeyFormat, item.GetRandId())

	createdAtAsString := item.GetCreatedAt().Format(FORMATTED_TIME)
	updatedAtAsString := item.GetUpdatedAt().Format(FORMATTED_TIME)

	item.SetCreatedAtString(createdAtAsString)
	item.SetUpdatedAtString(updatedAtAsString)

	itemInByte, errorMarshalJson := json.Marshal(item)
	if errorMarshalJson != nil {
		return &types.PaginationError{
			Err:     ERROR_MARSHAL_JSON,
			Details: errorMarshalJson.Error(),
		}
	}

	valueAsString := string(itemInByte)
	setRedis := cr.redisClient.Set(
		context.TODO(),
		key,
		valueAsString,
		INDIVIDUAL_KEY_TTL,
	)

	if setRedis.Err() != nil {
		return &types.PaginationError{
			Err:     REDIS_FATAL_ERROR,
			Details: setRedis.Err().Error(),
		}
	}

	fmt.Println(item.GetUUID())

	return nil
}

func (cr *ItemCacheType[T]) Del(item T) *types.PaginationError {
	key := fmt.Sprintf(cr.itemKeyFormat, item.GetRandId())

	deleteRedis := cr.redisClient.Del(
		context.TODO(),
		key,
	)

	if deleteRedis.Err() != nil {
		return &types.PaginationError{
			Err:     REDIS_FATAL_ERROR,
			Details: deleteRedis.Err().Error(),
		}
	}

	return nil
}
