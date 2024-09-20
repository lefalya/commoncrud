package commoncrud

import (
	"context"
	"log/slog"
	"time"

	"github.com/lefalya/commoncrud/interfaces"
	"github.com/lefalya/commoncrud/schema"
	"github.com/lefalya/commonlogger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoType[T interfaces.MongoItem] struct {
	logger           *slog.Logger
	collection       *mongo.Collection
	paginationFilter bson.A
}

func Mongo[T interfaces.MongoItem](logger *slog.Logger, collection *mongo.Collection) *MongoType[T] {
	return &MongoType[T]{
		logger:     logger,
		collection: collection,
	}
}

func bsonDToString(document bson.D) (string, error) {
	// Convert bson.D to extended JSON
	jsonData, err := bson.MarshalExtJSON(document, true, true)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func (mo *MongoType[T]) Create(item T) *schema.InternalError {
	createdAtStr := item.GetCreatedAt().Format(FORMATTED_TIME)
	updatedAtStr := item.GetUpdatedAt().Format(FORMATTED_TIME)

	item.SetCreatedAtString(createdAtStr)
	item.SetUpdatedAtString(updatedAtStr)

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

		return &schema.InternalError{
			Err:     MONGO_FATAL_ERROR,
			Details: errorCreate.Error(),
		}
	}

	return nil
}

func (mo *MongoType[T]) FindOne(randId string) (T, *schema.InternalError) {
	var nilItem T
	var item T
	filter := bson.D{{"item.randid", randId}}

	initializePointers(&item)

	findOneRes := mo.collection.FindOne(
		context.TODO(),
		filter,
	)
	findError := findOneRes.Decode(&item)
	if findError != nil {
		if findError == mongo.ErrNoDocuments {
			return nilItem, &schema.InternalError{
				Err:     DOCUMENT_NOT_FOUND,
				Details: "item not found!",
			}
		}

		return nilItem, &schema.InternalError{
			Err:     MONGO_FATAL_ERROR,
			Details: findError.Error(),
		}
	}

	parsedTime, errorParse := time.Parse(FORMATTED_TIME, item.GetCreatedAtString())
	if errorParse == nil {
		item.SetCreatedAt(parsedTime)
	}

	parsedTime, errorParse = time.Parse(FORMATTED_TIME, item.GetUpdatedAtString())
	if errorParse == nil {
		item.SetUpdatedAt(parsedTime)
	}

	return item, nil
}

func (mo *MongoType[T]) FindMany(
	filter bson.D,
	findOptions *options.FindOptions,
	pagination interfaces.Pagination[T],
	paginationParameters []string,
	processor interfaces.SeedProcessor[T],
	processorArgs ...interface{},
) ([]T, *schema.InternalError) {
	var results []T

	cursor, errorFindItems := mo.collection.Find(
		context.TODO(),
		filter,
		findOptions,
	)

	if errorFindItems != nil {
		return nil, &schema.InternalError{
			Err:     MONGO_FATAL_ERROR,
			Details: errorFindItems.Error(),
		}
	}

	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var item T

		errorDecode := cursor.Decode(&item)
		if errorDecode != nil {
			filterAsString, err := bsonDToString(filter)
			if err != nil {
				filterAsString = ""
			}

			commonlogger.LogError(
				mo.logger,
				FAILED_DECODE,
				errorDecode.Error(),
				"findMany.failed_decode",
				"filter", filterAsString,
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

		pagination.AddItem(paginationParameters, item)

		// During the seeding process, the processor functions solely as an item modifier/processor.
		// In contrast, during the fetching process, the processor also evaluates whether
		// each item meets the criteria to be included in the results.
		if processor != nil {
			processor(&item, processorArgs...)
		}

		results = append(results, item)
	}

	return results, nil
}

func (mo *MongoType[T]) Update(item T) *schema.InternalError {
	filter := bson.D{{"uuid", item.GetUUID()}}

	_, errorUpdate := mo.collection.UpdateOne(
		context.TODO(),
		filter,
		bson.D{{
			"$set",
			item,
		}},
	)

	if errorUpdate != nil {
		return &schema.InternalError{
			Err:     MONGO_FATAL_ERROR,
			Details: errorUpdate.Error(),
		}
	}

	return nil
}

func (mo *MongoType[T]) Delete(item T) *schema.InternalError {
	filter := bson.D{{"uuid", item.GetUUID()}}

	_, errorDelete := mo.collection.DeleteOne(
		context.TODO(),
		filter,
	)

	if errorDelete != nil {
		return &schema.InternalError{
			Err:     MONGO_FATAL_ERROR,
			Details: errorDelete.Error(),
		}
	}

	return nil
}

func (mo *MongoType[T]) SetPaginationFilter(filter bson.A) {
	mo.paginationFilter = filter
}

func (mo *MongoType[T]) GetPaginationFilter() bson.A {
	return mo.paginationFilter
}
