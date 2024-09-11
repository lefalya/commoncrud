package commonpagination

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestCreate(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	dummyItem := Student{
		Item: &Item{
			UUID:      uuid.New().String(),
			RandId:    RandId(),
			CreatedAt: time.Now().In(time.UTC),
			UpdatedAt: time.Now().In(time.UTC),
		},
		MongoItem: &MongoItem{
			ObjectId: primitive.NewObjectID(),
		},
		FirstName: "test",
		LastName:  "test again",
	}

	mt.Run("create success", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		mongo := Mongo[Student](logger, mt.Coll)
		errorCreate := mongo.Create(dummyItem)

		assert.Nil(t, errorCreate)
	})

	mt.Run("duplicate RandId", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   1,
			Code:    11000,
			Message: "duplicate key error",
		}))

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		mongo := Mongo[Student](logger, mt.Coll)
		errorCreate := mongo.Create(dummyItem)

		assert.Nil(t, errorCreate)
	})
}

func TestFind(t *testing.T) {
	currentTime := time.Now().In(time.UTC)

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	dummyItem := Student{
		Item: &Item{
			UUID:            uuid.New().String(),
			RandId:          RandId(),
			CreatedAtString: currentTime.Format(FORMATTED_TIME),
			UpdatedAtString: currentTime.Format(FORMATTED_TIME),
		},
		MongoItem: &MongoItem{
			ObjectId: primitive.NewObjectID(),
		},
		FirstName: "test",
		LastName:  "test again",
	}

	mt.Run("find MongoItem success", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.find", mtest.FirstBatch, bson.D{
			{"_id", dummyItem.ObjectId},
			{"uuid", dummyItem.UUID},
			{"randid", dummyItem.RandId},
			{"createdat", dummyItem.CreatedAtString},
			{"updatedat", dummyItem.UpdatedAtString},
			{"firstname", dummyItem.FirstName},
			{"lastname", dummyItem.LastName},
		}))

		mongo := Mongo[Student](logger, mt.Coll)

		itemFromDb, _ := mongo.Find(dummyItem.RandId)

		assert.NotNil(t, itemFromDb)
	})
}

func TestUpdate(t *testing.T) {
	currentTime := time.Now().In(time.UTC)

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	dummyItem := Student{
		Item: &Item{
			UUID:            uuid.New().String(),
			RandId:          RandId(),
			CreatedAtString: currentTime.Format(FORMATTED_TIME),
			UpdatedAtString: currentTime.Format(FORMATTED_TIME),
		},
		MongoItem: &MongoItem{
			ObjectId: primitive.NewObjectID(),
		},
		FirstName: "test",
		LastName:  "test again",
	}

	mt.Run("update success", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		mongo := Mongo[Student](logger, mt.Coll)
		errorUpdate := mongo.Update(dummyItem)

		assert.Nil(t, errorUpdate)
	})
}

func TestDelete(t *testing.T) {
	currentTime := time.Now().In(time.UTC)

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	dummyItem := Student{
		Item: &Item{
			UUID:            uuid.New().String(),
			RandId:          RandId(),
			CreatedAtString: currentTime.Format(FORMATTED_TIME),
			UpdatedAtString: currentTime.Format(FORMATTED_TIME),
		},
		MongoItem: &MongoItem{
			ObjectId: primitive.NewObjectID(),
		},
		FirstName: "test",
		LastName:  "test again",
	}

	mt.Run("delete success", func(mt *mtest.T) {

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		mongo := Mongo[Student](logger, mt.Coll)
		errorUpdate := mongo.Delete(dummyItem)

		assert.Nil(t, errorUpdate)
	})
}
