package commonpagination

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/lefalya/commonpagination/interfaces"
	mock_interfaces "github.com/lefalya/commonpagination/mocks"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestInjectMongo(t *testing.T) {
	type Injected[T interfaces.Item] struct {
		mongo interfaces.Mongo[T]
	}

	mongo := Mongo[Student](nil, nil)

	injected := Injected[Student]{
		mongo: mongo,
	}

	assert.NotNil(t, injected)
}

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

func TestFindOne(t *testing.T) {
	currentTime := time.Now().In(time.UTC)
	updatedTime := currentTime.Add(time.Hour)

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	dummyItem := Student{
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

		mongo := Mongo[Student](logger, mt.Coll)

		itemFromDb, _ := mongo.FindOne(dummyItem.RandId)

		assert.NotNil(t, itemFromDb)
		assert.Equal(t, currentTime, itemFromDb.CreatedAt)
		assert.Equal(t, currentTime.Day(), itemFromDb.CreatedAt.Day())
		assert.Equal(t, currentTime.Month(), itemFromDb.CreatedAt.Month())
		assert.Equal(t, currentTime.Year(), itemFromDb.CreatedAt.Year())
		assert.Equal(t, updatedTime, itemFromDb.UpdatedAt)
		assert.Equal(t, updatedTime.Day(), itemFromDb.UpdatedAt.Day())
		assert.Equal(t, updatedTime.Month(), itemFromDb.UpdatedAt.Month())
		assert.Equal(t, updatedTime.Year(), itemFromDb.UpdatedAt.Year())
	})
}

func TestUpdate(t *testing.T) {
	currentTime := time.Now().In(time.UTC)

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	dummyItem := Student{
		Item: &Item{
			UUID:      uuid.New().String(),
			RandId:    RandId(),
			CreatedAt: currentTime,
			UpdatedAt: currentTime,
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
			UUID:      uuid.New().String(),
			RandId:    RandId(),
			CreatedAt: currentTime,
			UpdatedAt: currentTime,
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

func TestFindMany(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	dummyItem1 := Student{
		Item: &Item{
			UUID:      uuid.New().String(),
			RandId:    RandId(),
			CreatedAt: time.Now().In(time.UTC),
			UpdatedAt: time.Now().In(time.UTC),
		},
		MongoItem: &MongoItem{
			ObjectId: primitive.NewObjectID(),
		},
		FirstName: "Fernando",
		LastName:  "Linblad",
	}

	dummyItem2 := Student{
		Item: &Item{
			UUID:      uuid.New().String(),
			RandId:    RandId(),
			CreatedAt: time.Now().In(time.UTC),
			UpdatedAt: time.Now().In(time.UTC),
		},
		MongoItem: &MongoItem{
			ObjectId: primitive.NewObjectID(),
		},
		FirstName: "Alice",
		LastName:  "Johnson",
	}

	dummyItem3 := Student{
		Item: &Item{
			UUID:      uuid.New().String(),
			RandId:    RandId(),
			CreatedAt: time.Now().In(time.UTC),
			UpdatedAt: time.Now().In(time.UTC),
		},
		MongoItem: &MongoItem{
			ObjectId: primitive.NewObjectID(),
		},
		FirstName: "Michael",
		LastName:  "Smith",
	}

	dummyItem4 := Student{
		Item: &Item{
			UUID:      uuid.New().String(),
			RandId:    RandId(),
			CreatedAt: time.Now().In(time.UTC),
			UpdatedAt: time.Now().In(time.UTC),
		},
		MongoItem: &MongoItem{
			ObjectId: primitive.NewObjectID(),
		},
		FirstName: "Laura",
		LastName:  "Martinez",
	}

	dummyItem5 := Student{
		Item: &Item{
			UUID:      uuid.New().String(),
			RandId:    RandId(),
			CreatedAt: time.Now().In(time.UTC),
			UpdatedAt: time.Now().In(time.UTC),
		},
		MongoItem: &MongoItem{
			ObjectId: primitive.NewObjectID(),
		},
		FirstName: "James",
		LastName:  "Wilson",
	}

	paginationKeyParameters := []string{uuid.New().String()}

	mt.Run("seed success with lastItem", func(mt *mtest.T) {
		dummyItem1Res := mtest.CreateCursorResponse(1, "test.seedPartial", mtest.FirstBatch, bson.D{
			{"_id", dummyItem1.ObjectId},
			{"uuid", dummyItem1.UUID},
			{"randid", dummyItem1.RandId},
			{"createdat", dummyItem1.CreatedAtString},
			{"updatedat", dummyItem1.UpdatedAtString},
			{"firstname", dummyItem1.FirstName},
			{"lastname", dummyItem1.LastName},
		})
		dummyItem2Res := mtest.CreateCursorResponse(1, "test.seedPartial", mtest.FirstBatch, bson.D{
			{"_id", dummyItem2.ObjectId},
			{"uuid", dummyItem2.UUID},
			{"randid", dummyItem2.RandId},
			{"createdat", dummyItem2.CreatedAtString},
			{"updatedat", dummyItem2.UpdatedAtString},
			{"firstname", dummyItem2.FirstName},
			{"lastname", dummyItem2.LastName},
		})
		dummyItem3Res := mtest.CreateCursorResponse(1, "test.seedPartial", mtest.FirstBatch, bson.D{
			{"_id", dummyItem3.ObjectId},
			{"uuid", dummyItem3.UUID},
			{"randid", dummyItem3.RandId},
			{"createdat", dummyItem3.CreatedAtString},
			{"updatedat", dummyItem3.UpdatedAtString},
			{"firstname", dummyItem3.FirstName},
			{"lastname", dummyItem3.LastName},
		})
		dummyItem4Res := mtest.CreateCursorResponse(1, "test.seedPartial", mtest.FirstBatch, bson.D{
			{"_id", dummyItem4.ObjectId},
			{"uuid", dummyItem4.UUID},
			{"randid", dummyItem4.RandId},
			{"createdat", dummyItem4.CreatedAtString},
			{"updatedat", dummyItem4.UpdatedAtString},
			{"firstname", dummyItem4.FirstName},
			{"lastname", dummyItem4.LastName},
		})
		dummyItem5Res := mtest.CreateCursorResponse(1, "test.seedPartial", mtest.FirstBatch, bson.D{
			{"_id", dummyItem5.ObjectId},
			{"uuid", dummyItem5.UUID},
			{"randid", dummyItem5.RandId},
			{"createdat", dummyItem5.CreatedAtString},
			{"updatedat", dummyItem5.UpdatedAtString},
			{"firstname", dummyItem5.FirstName},
			{"lastname", dummyItem5.LastName},
		})

		mt.AddMockResponses(
			dummyItem1Res,
			dummyItem2Res,
			dummyItem3Res,
			dummyItem4Res,
			dummyItem5Res,
		)
		killCursors := mtest.CreateCursorResponse(0, "test.seedPartial", mtest.NextBatch)
		mt.AddMockResponses(
			dummyItem1Res,
			dummyItem2Res,
			dummyItem3Res,
			dummyItem4Res,
			dummyItem5Res,
			killCursors,
		)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockPagiantion := mock_interfaces.NewMockPagination[Student](ctrl)
		mockPagiantion.EXPECT().AddItem(paginationKeyParameters, dummyItem1).Return(nil)
		mockPagiantion.EXPECT().AddItem(paginationKeyParameters, dummyItem2).Return(nil)
		mockPagiantion.EXPECT().AddItem(paginationKeyParameters, dummyItem3).Return(nil)
		mockPagiantion.EXPECT().AddItem(paginationKeyParameters, dummyItem4).Return(nil)
		mockPagiantion.EXPECT().AddItem(paginationKeyParameters, dummyItem5).Return(nil)

		// students, errorSeedPartially := mongo.SeedPartial()
	})
}
