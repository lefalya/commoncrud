package commoncrud

import (
	"context"
	"log/slog"
	"time"

	"github.com/lefalya/commoncrud/interfaces"
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

func (mo *MongoType[T]) Create(item T) *commonlogger.CommonError {
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

func (mo *MongoType[T]) FindOne(randId string) (T, *commonlogger.CommonError) {
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

func (mo *MongoType[T]) FindMany(filter bson.D, findOptions *options.FindOptions) (*mongo.Cursor, *commonlogger.CommonError) {

	cursor, errorFindItems := mo.collection.Find(
		context.TODO(),
		filter,
		findOptions,
	)

	if errorFindItems != nil {
		return nil, commonlogger.LogError(
			mo.logger,
			MONGO_FATAL_ERROR,
			errorFindItems.Error(),
			"findmany.find_mongodb_fatal_error",
		)
	}

	return cursor, nil
}

func (mo *MongoType[T]) Update(item T) *commonlogger.CommonError {
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

func (mo *MongoType[T]) SetPaginationFilter(filter bson.A) {
	mo.paginationFilter = filter
}

func (mo *MongoType[T]) GetPaginationFilter() bson.A {
	return mo.paginationFilter
}
