package commoncrud

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/lefalya/commoncrud/interfaces"
	"github.com/stretchr/testify/assert"
)

type Student struct {
	*Item      `bson:",inline"`
	*MongoItem `bson:",inline"`
	FirstName  string `bson:"firstname"`
	LastName   string `bson:"lastname"`
}

func TestInjectItemCache(t *testing.T) {
	type Injected[T interfaces.Item] struct {
		itemCache interfaces.ItemCache[T]
	}

	itemCache := ItemCache[Student]("", nil, nil)

	injected := Injected[Student]{
		itemCache: itemCache,
	}

	assert.NotNil(t, injected)
}

func TestGet(t *testing.T) {
	currentTime := time.Now().In(time.UTC)

	dummyItem := Student{
		Item: &Item{
			UUID:            uuid.New().String(),
			RandId:          RandId(),
			CreatedAtString: currentTime.Format(FORMATTED_TIME),
			UpdatedAtString: currentTime.Format(FORMATTED_TIME),
		},
		MongoItem: &MongoItem{},
		FirstName: "test",
		LastName:  "test again",
	}
	createdAtAsTime, _ := time.Parse(FORMATTED_TIME, dummyItem.GetCreatedAtString())
	updatedAtAsTime, _ := time.Parse(FORMATTED_TIME, dummyItem.GetUpdatedAtString())

	dummyItemKeyFormat := "student:%s"
	expectedKey := fmt.Sprintf(dummyItemKeyFormat, dummyItem.GetRandId())

	t.Run("successfully get item", func(t *testing.T) {
		jsonStringDummyItem, errorMarshal := json.Marshal(dummyItem)
		assert.Nil(t, errorMarshal)

		redisClient, mockRedis := redismock.NewClientMock()
		mockRedis.ExpectGet(expectedKey).SetVal(string(jsonStringDummyItem))
		mockRedis.ExpectExpire(expectedKey, INDIVIDUAL_KEY_TTL).SetVal(true)

		itemCache := ItemCache[Student](dummyItemKeyFormat, logger, redisClient)

		item, err := itemCache.Get(dummyItem.RandId)

		assert.NotNil(t, item)
		assert.Nil(t, err)
		assert.Equal(t, createdAtAsTime, item.CreatedAt)
		assert.Equal(t, createdAtAsTime.Day(), item.CreatedAt.Day())
		assert.Equal(t, createdAtAsTime.Month(), item.CreatedAt.Month())
		assert.Equal(t, createdAtAsTime.Year(), item.CreatedAt.Year())
		assert.Equal(t, updatedAtAsTime, item.UpdatedAt)
		assert.Equal(t, updatedAtAsTime.Day(), item.UpdatedAt.Day())
		assert.Equal(t, updatedAtAsTime.Month(), item.UpdatedAt.Month())
		assert.Equal(t, updatedAtAsTime.Year(), item.UpdatedAt.Year())
		assert.Equal(t, dummyItem.UUID, item.GetUUID())
		assert.Equal(t, dummyItem.RandId, item.GetRandId())
		assert.Equal(t, dummyItem.FirstName, item.FirstName)
		assert.Equal(t, dummyItem.LastName, item.LastName)
	})
}

func TestSet(t *testing.T) {

}

func TestDel(t *testing.T) {

}
