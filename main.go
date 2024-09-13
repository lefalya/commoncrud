package commonpagination

import (
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/lefalya/commonlogger"
	"github.com/lefalya/commonpagination/interfaces"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))
	// Go's reference time, which is Mon Jan 2 15:04:05 MST 2006
	FORMATTED_TIME = "2006-01-02T15:04:05.000000000Z"
	// Redis errors
	REDIS_FATAL_ERROR  = errors.New("(commonpagination) Redis fatal error")
	KEY_NOT_FOUND      = errors.New("(commonpagination) Key not found")
	ERROR_PARSE_JSON   = errors.New("(commonpagination) parse json fatal error!")
	ERROR_MARSHAL_JSON = errors.New("(commonpagination) error marshal json!")
	// Pagination errors
	TOO_MUCH_REFERENCES     = errors.New("(commonpagination) Too much references")
	NO_VALID_REFERENCES     = errors.New("(commonpagination) No valid references")
	LASTITEM_MUST_MONGOITEM = errors.New("(commonpagination) Last item must be in interfaces.MongoItem")
	NO_DATABASE_DEFINED     = errors.New("(commonpagination) Database not specified!")
	// MongoDB errors
	REFERENCE_NOT_FOUND = errors.New("(commonpagination) Reference not found")
	DOCUMENT_NOT_FOUND  = errors.New("(commonpagination) Document not found")
	MONGO_FATAL_ERROR   = errors.New("(commonpagination) MongoDB fatal error")
	DUPLICATE_RANDID    = errors.New("(commonpagination) Duplicate RandID")
	NO_OBJECTID_PRESENT = errors.New("(commonpagination) No objectId presents")
	FAILED_DECODE       = errors.New("(commonpagination) failed decode")
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
	args ...string,
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

func NewMongoItem[T interfaces.MongoItem](item T) T {
	currentTime := time.Now().In(time.UTC)

	initializePointers(&item)

	item.SetUUID()
	item.SetRandId()
	item.SetCreatedAt(currentTime)
	item.SetUpdatedAt(currentTime)
	item.SetObjectId()

	return item
}
