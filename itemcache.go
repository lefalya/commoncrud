package commoncrud

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/lefalya/commoncrud/interfaces"
	"github.com/lefalya/commonlogger"
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

func (cr *ItemCacheType[T]) Get(randId string) (T, *commonlogger.CommonError) {
	var nilItem T
	key := fmt.Sprintf(cr.itemKeyFormat, randId)

	result := cr.redisClient.Get(context.TODO(), key)

	if result.Err() != nil {
		if result.Err() == redis.Nil {
			return nilItem, commonlogger.LogError(
				cr.logger,
				KEY_NOT_FOUND,
				"key not found!",
				"get.key_not_found",
				"key", key,
			)
		}

		return nilItem, commonlogger.LogError(
			cr.logger,
			REDIS_FATAL_ERROR,
			result.Err().Error(),
			"get.fatal_error",
			"key", key,
		)
	}

	var item T
	errorUnmarshal := json.Unmarshal([]byte(result.Val()), &item)

	parsedTimeCreatedAt, _ := time.Parse(FORMATTED_TIME, item.GetCreatedAtString())
	parsedTimeUpdatedAt, _ := time.Parse(FORMATTED_TIME, item.GetUpdatedAtString())

	item.SetCreatedAt(parsedTimeCreatedAt)
	item.SetUpdatedAt(parsedTimeUpdatedAt)

	if errorUnmarshal != nil {
		return nilItem, commonlogger.LogError(
			cr.logger,
			ERROR_PARSE_JSON,
			errorUnmarshal.Error(),
			"get.error_parse_json",
			"key", key,
		)
	}

	setExpire := cr.redisClient.Expire(context.TODO(), key, INDIVIDUAL_KEY_TTL)
	if setExpire.Err() != nil {
		return nilItem, commonlogger.LogError(
			cr.logger,
			REDIS_FATAL_ERROR,
			setExpire.Err().Error(),
			"get.set_expire_fatal_error",
			"key", key,
		)
	}

	return item, nil
}

func (cr *ItemCacheType[T]) Set(item T) *commonlogger.CommonError {
	key := fmt.Sprintf(cr.itemKeyFormat, item.GetRandId())

	createdAtAsString := item.GetCreatedAt().Format(FORMATTED_TIME)
	updatedAtAsString := item.GetUpdatedAt().Format(FORMATTED_TIME)

	item.SetCreatedAtString(createdAtAsString)
	item.SetUpdatedAtString(updatedAtAsString)

	itemInByte, errorMarshalJson := json.Marshal(item)
	if errorMarshalJson != nil {
		return ItemLogHelper(
			cr.logger,
			ERROR_MARSHAL_JSON,
			errorMarshalJson.Error(),
			"set.json_marshal_error",
			item,
		)
	}

	valueAsString := string(itemInByte)
	setRedis := cr.redisClient.Set(
		context.TODO(),
		key,
		valueAsString,
		INDIVIDUAL_KEY_TTL,
	)

	if setRedis.Err() != nil {
		return ItemLogHelper(
			cr.logger,
			REDIS_FATAL_ERROR,
			setRedis.Err().Error(),
			"set.fatal_error",
			item,
		)
	}

	fmt.Println(item.GetUUID())

	return nil
}

func (cr *ItemCacheType[T]) Delete(item T) *commonlogger.CommonError {
	key := fmt.Sprintf(cr.itemKeyFormat, item.GetRandId())

	deleteRedis := cr.redisClient.Del(
		context.TODO(),
		key,
	)

	if deleteRedis.Err() != nil {
		return ItemLogHelper(
			cr.logger,
			REDIS_FATAL_ERROR,
			deleteRedis.Err().Error(),
			"delete.fatal_error",
			item,
		)
	}

	return nil
}
