package commoncrud

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/lefalya/commoncrud/interfaces"
	mock_interfaces "github.com/lefalya/commoncrud/mocks"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TestStructMongo struct {
	*Item      `bson:",inline"`
	*MongoItem `bson:",inline"`
	FirstName  string `bson:"firstname"`
	LastName   string `bson:"lastname"`
}

func TestInjectMongo(t *testing.T) {
	type Injected[T interfaces.Item] struct {
		mongo interfaces.Mongo[T]
	}

	mongo := Mongo[TestStructMongo](nil, nil)

	injected := Injected[TestStructMongo]{
		mongo: mongo,
	}

	assert.NotNil(t, injected)
}

func TestCreate(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	dummyItem := TestStructMongo{
		FirstName: "Walter",
		LastName:  "White",
	}

	dummyItem = NewMongoItem(dummyItem)

	mt.Run("create success", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		mongo := Mongo[TestStructMongo](logger, mt.Coll)
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

		mongo := Mongo[TestStructMongo](logger, mt.Coll)
		errorCreate := mongo.Create(dummyItem)

		assert.Nil(t, errorCreate)
	})
	mt.Run("create failure - MONGO_FATAL_ERROR", func(mt *mtest.T) {
		mt.AddMockResponses(bson.D{{"ok", 0}})

		mongo := Mongo[TestStructMongo](logger, mt.Coll)
		errorCreate := mongo.Create(dummyItem)

		assert.NotNil(t, errorCreate)
	})
}

func TestFindOne(t *testing.T) {
	currentTime := time.Now().In(time.UTC)
	updatedTime := currentTime.Add(time.Hour)

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	dummyItem := TestStructMongo{
		Item: &Item{
			UUID:            uuid.New().String(),
			RandId:          RandId(),
			CreatedAtString: currentTime.Format(FORMATTED_TIME),
			UpdatedAtString: updatedTime.Format(FORMATTED_TIME),
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

		mongo := Mongo[TestStructMongo](logger, mt.Coll)

		item, _ := mongo.FindOne(dummyItem.RandId)

		assert.NotNil(t, item)
		assert.Equal(t, currentTime, item.CreatedAt)
		assert.Equal(t, currentTime.Day(), item.CreatedAt.Day())
		assert.Equal(t, currentTime.Month(), item.CreatedAt.Month())
		assert.Equal(t, currentTime.Year(), item.CreatedAt.Year())
		assert.Equal(t, updatedTime, item.UpdatedAt)
		assert.Equal(t, updatedTime.Day(), item.UpdatedAt.Day())
		assert.Equal(t, updatedTime.Month(), item.UpdatedAt.Month())
		assert.Equal(t, updatedTime.Year(), item.UpdatedAt.Year())
		assert.Equal(t, dummyItem.FirstName, item.FirstName)
		assert.Equal(t, dummyItem.LastName, item.LastName)
	})

	mt.Run("item not found", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(0, "test.find", mtest.FirstBatch))

		mongo := Mongo[TestStructMongo](logger, mt.Coll)
		item, errorFind := mongo.FindOne(RandId())

		assert.NotNil(t, errorFind)
		assert.NotNil(t, item)
		assert.Equal(t, errorFind.Err, DOCUMENT_NOT_FOUND)
		assert.NotNil(t, errorFind.Details)
	})
	mt.Run("mongo fatal error", func(mt *mtest.T) {
		mt.AddMockResponses(bson.D{{"ok", 0}})

		mongo := Mongo[TestStructMongo](logger, mt.Coll)
		item, errorFind := mongo.FindOne(RandId())

		assert.NotNil(t, errorFind)
		assert.NotNil(t, item)
		assert.Equal(t, errorFind.Err, MONGO_FATAL_ERROR)
		assert.NotNil(t, errorFind.Details)
	})
	mt.Run("failed to parse CreatedAt time", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.find", mtest.FirstBatch, bson.D{
			{"_id", dummyItem.ObjectId},
			{"uuid", dummyItem.UUID},
			{"randid", dummyItem.RandId},
			{"createdat", "Mon, 4 July 2024 12:23:34"},
			{"updatedat", dummyItem.UpdatedAtString},
			{"firstname", dummyItem.FirstName},
			{"lastname", dummyItem.LastName},
		}))

		mongo := Mongo[TestStructMongo](logger, mt.Coll)
		item, errorFind := mongo.FindOne(RandId())

		assert.NotNil(t, item)
		assert.Nil(t, errorFind)
		assert.Equal(t, time.Time{}, item.CreatedAt)
		assert.Equal(t, updatedTime, item.UpdatedAt)
		assert.Equal(t, updatedTime.Day(), item.UpdatedAt.Day())
		assert.Equal(t, updatedTime.Month(), item.UpdatedAt.Month())
		assert.Equal(t, updatedTime.Year(), item.UpdatedAt.Year())
		assert.Equal(t, dummyItem.FirstName, item.FirstName)
		assert.Equal(t, dummyItem.LastName, item.LastName)
	})
	mt.Run("failed to parse UpdatedAt time", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "test.find", mtest.FirstBatch, bson.D{
			{"_id", dummyItem.ObjectId},
			{"uuid", dummyItem.UUID},
			{"randid", dummyItem.RandId},
			{"createdat", dummyItem.CreatedAtString},
			{"updatedat", "Mon, 4 July 2024 12:23:34"},
			{"firstname", dummyItem.FirstName},
			{"lastname", dummyItem.LastName},
		}))

		mongo := Mongo[TestStructMongo](logger, mt.Coll)
		item, errorFind := mongo.FindOne(RandId())

		assert.NotNil(t, item)
		assert.Nil(t, errorFind)
		assert.Equal(t, time.Time{}, item.UpdatedAt)
		assert.Equal(t, currentTime, item.CreatedAt)
		assert.Equal(t, currentTime.Day(), item.CreatedAt.Day())
		assert.Equal(t, currentTime.Month(), item.CreatedAt.Month())
		assert.Equal(t, currentTime.Year(), item.CreatedAt.Year())
		assert.Equal(t, dummyItem.FirstName, item.FirstName)
		assert.Equal(t, dummyItem.LastName, item.LastName)
	})
}

func TestUpdate(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	dummyItem := TestStructMongo{
		FirstName: "Walter",
		LastName:  "White",
	}

	dummyItem = NewMongoItem(dummyItem)

	mt.Run("update success", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		mongo := Mongo[TestStructMongo](logger, mt.Coll)
		errorUpdate := mongo.Update(dummyItem)

		assert.Nil(t, errorUpdate)
	})
	mt.Run("fatal error", func(mt *mtest.T) {
		mt.AddMockResponses(bson.D{{"ok", 0}})

		mongo := Mongo[TestStructMongo](logger, mt.Coll)
		errorUpdate := mongo.Update(dummyItem)

		assert.NotNil(t, errorUpdate)
		assert.Equal(t, MONGO_FATAL_ERROR, errorUpdate.Err)
		assert.NotNil(t, errorUpdate.Details)
	})
}

func TestDelete(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	dummyItem := TestStructMongo{
		FirstName: "Walter",
		LastName:  "White",
	}

	dummyItem = NewMongoItem(dummyItem)
	mt.Run("delete success", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		mongo := Mongo[TestStructMongo](logger, mt.Coll)
		errorUpdate := mongo.Delete(dummyItem)

		assert.Nil(t, errorUpdate)
	})
	mt.Run("fatal error", func(mt *mtest.T) {
		mt.AddMockResponses(bson.D{{"ok", 0}})

		mongo := Mongo[TestStructMongo](logger, mt.Coll)
		errorUpdate := mongo.Delete(dummyItem)

		assert.NotNil(t, errorUpdate)
		assert.Equal(t, MONGO_FATAL_ERROR, errorUpdate.Err)
		assert.NotNil(t, errorUpdate.Details)
	})
}

func StringProcess(s *string) {
	*s = *s + " proceseed"
}

func TestFindMany(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	dummyItem1 := TestStructMongo{
		Item: &Item{
			UUID:            uuid.New().String(),
			RandId:          RandId(),
			CreatedAtString: time.Now().In(time.UTC).String(),
			UpdatedAtString: time.Now().In(time.UTC).String(),
		},
		MongoItem: &MongoItem{
			ObjectId: primitive.NewObjectID(),
		},
		FirstName: "Fernando",
		LastName:  "Linblad",
	}

	dummyItem2 := TestStructMongo{
		Item: &Item{
			UUID:            uuid.New().String(),
			RandId:          RandId(),
			CreatedAtString: time.Now().In(time.UTC).String(),
			UpdatedAtString: time.Now().In(time.UTC).String(),
		},
		MongoItem: &MongoItem{
			ObjectId: primitive.NewObjectID(),
		},
		FirstName: "Alice",
		LastName:  "Johnson",
	}

	dummyItem3 := TestStructMongo{
		Item: &Item{
			UUID:            uuid.New().String(),
			RandId:          RandId(),
			CreatedAtString: time.Now().In(time.UTC).String(),
			UpdatedAtString: time.Now().In(time.UTC).String(),
		},
		MongoItem: &MongoItem{
			ObjectId: primitive.NewObjectID(),
		},
		FirstName: "Michael",
		LastName:  "Smith",
	}

	dummyItem4 := TestStructMongo{
		Item: &Item{
			UUID:            uuid.New().String(),
			RandId:          RandId(),
			CreatedAtString: time.Now().In(time.UTC).String(),
			UpdatedAtString: time.Now().In(time.UTC).String(),
		},
		MongoItem: &MongoItem{
			ObjectId: primitive.NewObjectID(),
		},
		FirstName: "Laura",
		LastName:  "Martinez",
	}

	dummyItem5 := TestStructMongo{
		Item: &Item{
			UUID:            uuid.New().String(),
			RandId:          RandId(),
			CreatedAtString: time.Now().In(time.UTC).String(),
			UpdatedAtString: time.Now().In(time.UTC).String(),
		},
		MongoItem: &MongoItem{
			ObjectId: primitive.NewObjectID(),
		},
		FirstName: "James",
		LastName:  "Wilson",
	}

	pagParams := []string{"Volkswagen", "SUV"}

	mt.Run("success with lastItem & processor", func(mt *mtest.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPagination := mock_interfaces.NewMockPagination[TestStructMongo](ctrl)
		mockPagination.EXPECT().AddItem(pagParams, dummyItem1)
		mockPagination.EXPECT().AddItem(pagParams, dummyItem2)
		mockPagination.EXPECT().AddItem(pagParams, dummyItem3)
		mockPagination.EXPECT().AddItem(pagParams, dummyItem4)
		mockPagination.EXPECT().AddItem(pagParams, dummyItem5)

		dummyItem1Res := mtest.CreateCursorResponse(1, "test.seedPartial", mtest.FirstBatch, bson.D{
			{"_id", dummyItem1.ObjectId},
			{"uuid", dummyItem1.UUID},
			{"randid", dummyItem1.RandId},
			{"createdat", dummyItem1.CreatedAtString},
			{"updatedat", dummyItem1.UpdatedAtString},
			{"firstname", dummyItem1.FirstName},
			{"lastname", dummyItem1.LastName},
		})
		dummyItem2Res := mtest.CreateCursorResponse(1, "test.seedPartial", mtest.NextBatch, bson.D{
			{"_id", dummyItem2.ObjectId},
			{"uuid", dummyItem2.UUID},
			{"randid", dummyItem2.RandId},
			{"createdat", dummyItem2.CreatedAtString},
			{"updatedat", dummyItem2.UpdatedAtString},
			{"firstname", dummyItem2.FirstName},
			{"lastname", dummyItem2.LastName},
		})
		dummyItem3Res := mtest.CreateCursorResponse(1, "test.seedPartial", mtest.NextBatch, bson.D{
			{"_id", dummyItem3.ObjectId},
			{"uuid", dummyItem3.UUID},
			{"randid", dummyItem3.RandId},
			{"createdat", dummyItem3.CreatedAtString},
			{"updatedat", dummyItem3.UpdatedAtString},
			{"firstname", dummyItem3.FirstName},
			{"lastname", dummyItem3.LastName},
		})
		dummyItem4Res := mtest.CreateCursorResponse(1, "test.seedPartial", mtest.NextBatch, bson.D{
			{"_id", dummyItem4.ObjectId},
			{"uuid", dummyItem4.UUID},
			{"randid", dummyItem4.RandId},
			{"createdat", dummyItem4.CreatedAtString},
			{"updatedat", dummyItem4.UpdatedAtString},
			{"firstname", dummyItem4.FirstName},
			{"lastname", dummyItem4.LastName},
		})
		dummyItem5Res := mtest.CreateCursorResponse(1, "test.seedPartial", mtest.NextBatch, bson.D{
			{"_id", dummyItem5.ObjectId},
			{"uuid", dummyItem5.UUID},
			{"randid", dummyItem5.RandId},
			{"createdat", dummyItem5.CreatedAtString},
			{"updatedat", dummyItem5.UpdatedAtString},
			{"firstname", dummyItem5.FirstName},
			{"lastname", dummyItem5.LastName},
		})
		killCursors := mtest.CreateCursorResponse(0, "test.seedPartial", mtest.NextBatch)
		mt.AddMockResponses(
			dummyItem1Res,
			dummyItem2Res,
			dummyItem3Res,
			dummyItem4Res,
			dummyItem5Res,
			killCursors,
		)

		filter := bson.D{}
		findOptions := options.Find()
		findOptions.SetSort(bson.D{{"_id", -1}})
		findOptions.SetLimit(10)

		mongo := Mongo[TestStructMongo](logger, mt.Coll)
		items, errorFindMany := mongo.FindMany(
			filter,
			findOptions,
			nil,
			pagParams,
			func(item *TestStructMongo) {
				StringProcess(&item.UUID)
				fmt.Println(item.UUID)
			},
		)

		assert.Nil(t, errorFindMany)
		assert.NotNil(t, items)

		assert.Equal(t, 5, len(items))
		assert.Equal(t, items[0].FirstName, dummyItem1.FirstName)
		assert.Equal(t, items[0].LastName, dummyItem1.LastName)
		assert.Equal(t, items[1].FirstName, dummyItem2.FirstName)
		assert.Equal(t, items[1].LastName, dummyItem2.LastName)
		assert.Equal(t, items[2].FirstName, dummyItem3.FirstName)
		assert.Equal(t, items[2].LastName, dummyItem3.LastName)
		assert.Equal(t, items[3].FirstName, dummyItem4.FirstName)
		assert.Equal(t, items[3].LastName, dummyItem4.LastName)
		assert.Equal(t, items[4].FirstName, dummyItem5.FirstName)
		assert.Equal(t, items[4].LastName, dummyItem5.LastName)
	})
}
