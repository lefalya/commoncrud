package commonpagination

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lefalya/commonlogger"
	"github.com/lefalya/commonpagination/interfaces"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	DAY                        = 24 * time.Hour
	INDIVIDUAL_KEY_TTL         = DAY * 7
	SORTED_SET_TTL             = DAY * 2
	MAXIMUM_AMOUNT_REFERENCES  = 5
	RANDID_LENGTH              = 16
	MONGO_DUPLICATE_ERROR_CODE = 11000
)

var (
	// Go's reference time, which is Mon Jan 2 15:04:05 MST 2006
	FORMATTED_TIME = "2006-01-02T15:04:05.000000000Z"
	// Redis errors
	REDIS_FATAL_ERROR  = errors.New("(commoncrud) Redis fatal error")
	KEY_NOT_FOUND      = errors.New("(commoncrud) Key not found")
	ERROR_PARSE_JSON   = errors.New("(commoncrud) parse json fatal error!")
	ERROR_MARSHAL_JSON = errors.New("(commoncrud) error marshal json!")
	// Pagination errors
	TOO_MUCH_REFERENCES = errors.New("(commoncrud) Too much references")
	NO_VALID_REFERENCES = errors.New("(commoncrud) No valid references")
	// MongoDB errors
	REFERENCE_NOT_FOUND = errors.New("(commoncrud) Reference not found")
	DOCUMENT_NOT_FOUND  = errors.New("(commoncrud) Document not found")
	MONGO_FATAL_ERROR   = errors.New("(commoncrud) MongoDB fatal error")
	DUPLICATE_RANDID    = errors.New("(commoncrud) Duplicate RandID")
	NO_OBJECTID_PRESENT = errors.New("(commoncrud) No objectId presents")
	FAILED_DECODE       = errors.New("(commoncrud) failed decode")
	INVALID_HEX         = errors.New("(commonpagination) Invalid hex: fail to convert hex to ObjectId")
)

func concatKey(keyFormat string, parameters []string) string {
	args := make([]interface{}, len(parameters))
	for i, v := range parameters {
		args[i] = v
	}

	return fmt.Sprintf(keyFormat, args...)
}

func RandId() string {
	// Define the characters that can be used in the random string
	characters := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Initialize an empty string to store the result
	result := make([]byte, RANDID_LENGTH)

	// Generate random characters for the string
	for i := 0; i < RANDID_LENGTH; i++ {
		result[i] = characters[rand.Intn(len(characters))]
	}

	return string(result)
}

type Item struct {
	UUID            string    `bson:"uuid"`
	RandId          string    `bson:"randid"`
	CreatedAt       time.Time `json:"-" bson:"-"`
	UpdatedAt       time.Time `json:"-" bson:"-"`
	CreatedAtString string    `bson:"createdat"`
	UpdatedAtString string    `bson:"updatedat"`
}

func (i *Item) SetUUID() {
	i.UUID = uuid.New().String()
}

func (i *Item) GetUUID() string {
	return i.UUID
}

func (i *Item) SetRandId() {
	i.RandId = RandId()
}

func (i *Item) GetRandId() string {
	return i.RandId
}

func (i *Item) SetCreatedAt(time time.Time) {
	i.CreatedAt = time
}

func (i *Item) SetUpdatedAt(time time.Time) {
	i.UpdatedAt = time
}

func (i *Item) GetCreatedAt() time.Time {
	return i.CreatedAt
}

func (i *Item) GetUpdatedAt() time.Time {
	return i.UpdatedAt
}

func (i *Item) SetCreatedAtString(timeString string) {
	i.CreatedAtString = timeString
}

func (i *Item) SetUpdatedAtString(timeString string) {
	i.UpdatedAtString = timeString
}

func (i *Item) GetCreatedAtString() string {
	return i.CreatedAtString
}

func (i *Item) GetUpdatedAtString() string {
	return i.UpdatedAtString
}

func ItemLogHelper[T interfaces.Item](
	logger *slog.Logger,
	returnedError error,
	errorDetail string,
	context string,
	item T,
) *commonlogger.CommonError {
	return commonlogger.LogError(
		logger,
		returnedError,
		errorDetail,
		context,
		"UUID", item.GetUUID(),
		"RandId", item.GetRandId(),
		"CreatedAt", item.GetCreatedAt().String(),
		"UpdatedAt", item.GetUpdatedAt().String(),
	)
}

type MongoItem struct {
	ObjectId primitive.ObjectID `bson:"_id"`
}

func (m *MongoItem) SetObjectId() {
	m.ObjectId = primitive.NewObjectID()
}

func (m *MongoItem) GetObjectId() primitive.ObjectID {
	return m.ObjectId
}

type ItemCacheType[T interfaces.Item] struct {
	itemKeyFormat string
	logger        *slog.Logger
	redisClient   *redis.Client
}

func ItemCache[T interfaces.Item](keyFormat string, logger *slog.Logger, redisClient *redis.Client) *ItemCacheType[T] {
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

type PaginationType[T interfaces.Item] struct {
	pagKeyFormat  string
	itemKeyFormat string
	itemPerPage   int64
	logger        *slog.Logger
	redisClient   *redis.Client
	itemCache     interfaces.ItemCache[T]
}

func Pagination[T interfaces.Item](
	pagKeyFormat string,
	itemKeyFormat string,
	itemPerPage int64,
	logger *slog.Logger,
	redisClient *redis.Client,
) *PaginationType[T] {
	itemCache := &ItemCacheType[T]{
		logger:        logger,
		redisClient:   redisClient,
		itemKeyFormat: itemKeyFormat,
	}

	return &PaginationType[T]{
		pagKeyFormat:  pagKeyFormat,
		itemKeyFormat: itemKeyFormat,
		itemPerPage:   itemPerPage,
		logger:        logger,
		redisClient:   redisClient,
		itemCache:     itemCache,
	}
}

func (pg *PaginationType[T]) ItemPerPage() int64 {
	return pg.itemPerPage
}

func (pg *PaginationType[T]) AddItem(pagKeyParams []string, item T) *commonlogger.CommonError {
	errorArgs := []string{
		"pagKeyFormat", pg.pagKeyFormat,
		"pagKeyParams", strings.Join(pagKeyParams, ","),
		"score", fmt.Sprintf("%f", float64(item.GetCreatedAt().Unix())),
		"member", item.GetRandId(),
	}

	key := concatKey(pg.pagKeyFormat, pagKeyParams)

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
		return commonlogger.LogError(
			pg.logger,
			REDIS_FATAL_ERROR,
			setSortedSet.Err().Error(),
			"additem.zadd_redis_fatal_error",
			errorArgs...,
		)
	}

	setExpire := pg.redisClient.Expire(
		context.TODO(),
		key,
		SORTED_SET_TTL,
	)

	if setExpire.Err() != nil {
		return commonlogger.LogError(
			pg.logger,
			REDIS_FATAL_ERROR,
			setExpire.Err().Error(),
			"additem.setexpire_redis_fatal_error",
			errorArgs...,
		)
	}

	return nil
}

func (pg *PaginationType[T]) RemoveItem(pagKeyParams []string, item T) *commonlogger.CommonError {
	key := concatKey(pg.pagKeyFormat, pagKeyParams)

	removeItemSortedSet := pg.redisClient.ZRem(context.TODO(), key, item.GetRandId())

	if removeItemSortedSet.Err() != nil {
		return commonlogger.LogError(
			pg.logger,
			REDIS_FATAL_ERROR,
			removeItemSortedSet.Err().Error(),
			"removeitem.zrem_redis_fatal_error",
			"pagKeyFormat", pg.pagKeyFormat,
			"pagKeyParams", strings.Join(pagKeyParams, ","),
			"member", item.GetRandId(),
		)
	}

	return nil
}

func (pg *PaginationType[T]) TotalItem(pagKeyParams []string, item T) *commonlogger.CommonError {
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

func (pg *PaginationType[T]) FetchLinked(
	pagKeyParams []string,
	references []string,
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

	stop = start + pg.itemPerPage - 1

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

			processor(item, items, pg.redisClient, processorArgs...)
		}
	}

	return items, nil
}

type MongoType[T interfaces.MongoItem] struct {
	logger     *slog.Logger
	collection *mongo.Collection
	itemCache  interfaces.ItemCache[T]
	pagination interfaces.Pagination[T]
}

func Mongo[T interfaces.MongoItem](logger *slog.Logger, collection *mongo.Collection) *MongoType[T] {
	return &MongoType[T]{
		logger:     logger,
		collection: collection,
	}
}

func (mo *MongoType[T]) WithPagination(itemCache interfaces.ItemCache[T], pagination interfaces.Pagination[T]) {
	mo.itemCache = itemCache
	mo.pagination = pagination
}

func (mo *MongoType[T]) Create(item T) *commonlogger.CommonError {

	_, errorCreate := mo.collection.InsertOne(
		context.TODO(),
		item,
	)

	if errorCreate != nil {
		if writeException, ok := errorCreate.(mongo.WriteException); ok {
			ItemLogHelper(
				mo.logger,
				DUPLICATE_RANDID,
				errorCreate.Error(),
				"create.duplicate_randid",
				item,
			)

			for _, werror := range writeException.WriteErrors {
				if werror.Code == MONGO_DUPLICATE_ERROR_CODE {
					item.SetRandId()
					mo.Create(item)

					return nil
				}
			}
		}
		return ItemLogHelper(
			mo.logger,
			MONGO_FATAL_ERROR,
			errorCreate.Error(),
			"create.mongo_fatal_error",
			item,
		)
	}

	return nil
}

func initializePointers(item interface{}) {
	value := reflect.ValueOf(item).Elem()

	// Iterate through the fields of the struct
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)

		// Check if the field is a pointer and is nil
		if field.Kind() == reflect.Ptr && field.IsNil() {
			// Allocate a new value for the pointer and set it
			field.Set(reflect.New(field.Type().Elem()))
		}
	}
}

func (mo *MongoType[T]) Find(randId string) (T, *commonlogger.CommonError) {
	var nilItem T

	filter := bson.D{{"randId", randId}}

	var item T
	initializePointers(&item)
	findOneRes := mo.collection.FindOne(
		context.TODO(),
		filter,
	)

	findError := findOneRes.Decode(&item)

	if findError != nil {
		if findError == mongo.ErrNoDocuments {
			return nilItem, commonlogger.LogError(
				mo.logger,
				DOCUMENT_NOT_FOUND,
				"document not found!",
				"find.document_not_found",
				"randId", randId,
			)
		}

		return nilItem, commonlogger.LogError(
			mo.logger,
			MONGO_FATAL_ERROR,
			findError.Error(),
			"find.mongodb_fatal_error",
			"randId", randId,
		)
	}

	return item, nil
}

func (mo *MongoType[T]) Update(item T) *commonlogger.CommonError {
	filter := bson.D{{"uuid", item.GetUUID()}}

	// updateList, _ := structToBSON(item)

	_, errorUpdate := mo.collection.UpdateOne(
		context.TODO(),
		filter,
		bson.D{{
			"$set",
			item,
		}},
	)

	if errorUpdate != nil {
		return ItemLogHelper(
			mo.logger,
			MONGO_FATAL_ERROR,
			errorUpdate.Error(),
			"update.mongodb_fatal_error",
			item,
		)
	}

	return nil
}

func (mo *MongoType[T]) Delete(item T) *commonlogger.CommonError {
	filter := bson.D{{"uuid", item.GetUUID()}}

	_, errorDelete := mo.collection.DeleteOne(
		context.TODO(),
		filter,
	)

	if errorDelete != nil {
		return ItemLogHelper(
			mo.logger,
			MONGO_FATAL_ERROR,
			errorDelete.Error(),
			"delete.mongodb_fatal_error",
			item,
		)
	}

	return nil
}

// func (mo *MongoType[T]) SeedLinked(
// 	paginationKeyParameters []string,
// 	lastItem T,
// 	paginationFilter bson.A,
// 	processor interfaces.MongoProcesor[T],
// 	processorArgs ...interface{},
// ) ([]T, *commonlogger.CommonError) {
// 	errorArgs := []string{}

// 	var cursor *mongo.Cursor
// 	var filter bson.D
// 	var result []T

// 	findOptions := options.Find()
// 	findOptions.SetSort(bson.D{{"_id", -1}})
// 	findOptions.SetLimit(mo.pagination.ItemPerPage())

// 	if lastItem != nil {
// 		filter = bson.D{
// 			{
// 				Key: "$and",
// 				Value: append(
// 					paginationFilter,
// 					bson.A{
// 						bson.D{
// 							{
// 								Key: "_id",
// 								Value: bson.D{
// 									{
// 										Key:   "$lt",
// 										Value: lastItem.GetObjectId(),
// 									},
// 								},
// 							},
// 						},
// 					}...,
// 				),
// 			},
// 		}
// 	} else {
// 		filter = bson.D{
// 			{Key: "$and",
// 				Value: append(
// 					paginationFilter,
// 				),
// 			},
// 		}
// 	}

// 	var errorFindItems error

// 	cursor, errorFindItems = mo.collection.Find(
// 		context.TODO(),
// 		filter,
// 		findOptions,
// 	)

// 	if errorFindItems != nil {

// 		return nil, commonlogger.LogError(
// 			mo.logger,
// 			MONGO_FATAL_ERROR,
// 			errorFindItems.Error(),
// 			"seedlinkedpagination.find_mongodb_fatal_error",
// 			errorArgs...,
// 		)
// 	}

// 	defer cursor.Close(context.TODO())

// 	for cursor.Next(context.TODO()) {

// 		var item T
// 		errorDecode := cursor.Decode(&item)

// 		if errorDecode != nil {

// 			commonlogger.LogError(
// 				mo.logger,
// 				FAILED_DECODE,
// 				errorDecode.Error(),
// 				"seedlinkedpagination.failed_decode",
// 				errorArgs...,
// 			)

// 			continue
// 		}

// 		processor(
// 			&item,
// 			ms.redisClient,
// 			ms.mongoCollection,
// 			processorArgs...,
// 		)

// 		ms.linkedPagination.AddItem()
// 	}

// 	return nil
// }
