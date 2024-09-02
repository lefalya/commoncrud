package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/lefalya/commonlogger"
	loggerSchema "github.com/lefalya/commonlogger/schema"
	"github.com/redis/go-redis/v9"
	"github.com/zeebo/assert"
)

const (
	INDIVIDUAL_SUBMISSION_KEY         = "submission:%s"
	SORTED_SET_PARTICIPANT_SUBMISSION = "campaign:%s:submission:%s"
)

var (
	logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))
)

type Submission struct {
	UUID            string
	CampaignUUID    string
	ParticipantUUID string
	Value           string
	CreatedAt       time.Time
}

func LogErrorSubmission(
	logger *slog.Logger,
	errorSubject error,
	errorDetail string,
	context string,
	submission any,
) *loggerSchema.CommonError {

	submissionAs := submission.(Submission)

	return commonlogger.LogError(
		logger,
		errorSubject,
		errorDetail,
		context,
		"uuid", submissionAs.UUID,
		"caption", submissionAs.Value,
	)
}

func TestPaginateAddItem(t *testing.T) {

	dummyCampaignUUID := uuid.New().String()
	dummyParticipantUUID := uuid.New().String()
	currentTime := time.Now()

	dummyItem := Submission{
		UUID:            uuid.New().String(),
		ParticipantUUID: dummyParticipantUUID,
		CampaignUUID:    dummyCampaignUUID,
		Value:           "dummy submission",
		CreatedAt:       currentTime,
	}

	expectedZMember := redis.Z{
		Score:  float64(dummyItem.CreatedAt.Unix()),
		Member: dummyItem.UUID,
	}

	expectedKey := fmt.Sprintf(SORTED_SET_PARTICIPANT_SUBMISSION, dummyCampaignUUID, dummyParticipantUUID)

	t.Run("successfull run sorted set", func(t *testing.T) {

		redisClient, mockRedis := redismock.NewClientMock()

		mockRedis.ExpectZAdd(
			expectedKey,
			expectedZMember,
		).SetVal(1)

		mockRedis.ExpectExpire(
			expectedKey,
			SORTED_SET_TTL,
		).SetVal(true)

		paginate := NewLinkedPagination[Submission](
			logger,
			redisClient,
			nil,
		)

		errorSetSortedSet := paginate.AddItem(
			SORTED_SET_PARTICIPANT_SUBMISSION,
			float64(dummyItem.CreatedAt.Unix()),
			dummyItem.UUID,
			&dummyItem,
			"testsortedset",
			LogErrorSubmission,
			dummyCampaignUUID,
			dummyParticipantUUID,
		)
		assert.Nil(t, errorSetSortedSet)
	})

	t.Run("fatal error on ZAdd", func(t *testing.T) {

		redisClient, mockRedis := redismock.NewClientMock()

		mockRedis.ExpectZAdd(
			expectedKey,
			expectedZMember,
		).SetErr(errors.New("fatal error: connection lost"))

		paginate := NewLinkedPagination[Submission](
			logger,
			redisClient,
			nil,
		)

		errorSetSortedSet := paginate.AddItem(
			SORTED_SET_PARTICIPANT_SUBMISSION,
			float64(dummyItem.CreatedAt.Unix()),
			dummyItem.UUID,
			&dummyItem,
			"testsortedset",
			LogErrorSubmission,
			dummyCampaignUUID,
			dummyParticipantUUID,
		)
		assert.NotNil(t, errorSetSortedSet)
		assert.Equal(t, REDIS_FATAL_ERROR, errorSetSortedSet.Err)
		assert.Equal(t, "testsortedset.set_sorted_set_REDIS_FATAL_ERROR", errorSetSortedSet.Context)
	})

	t.Run("fatal error on Expire", func(t *testing.T) {

		redisClient, mockRedis := redismock.NewClientMock()

		mockRedis.ExpectZAdd(
			expectedKey,
			expectedZMember,
		).SetVal(1)

		mockRedis.ExpectExpire(
			expectedKey,
			SORTED_SET_TTL,
		).SetErr(errors.New("fatal error: connection lost"))

		paginate := NewLinkedPagination[Submission](
			logger,
			redisClient,
			nil,
		)

		errorSetSortedSet := paginate.AddItem(
			SORTED_SET_PARTICIPANT_SUBMISSION,
			float64(dummyItem.CreatedAt.Unix()),
			dummyItem.UUID,
			&dummyItem,
			"testsortedset",
			LogErrorSubmission,
			dummyCampaignUUID,
			dummyParticipantUUID,
		)
		assert.NotNil(t, errorSetSortedSet)
		assert.Equal(t, REDIS_FATAL_ERROR, errorSetSortedSet.Err)
		assert.Equal(t, "testsortedset.set_sorted_set_expire_REDIS_FATAL_ERROR", errorSetSortedSet.Context)
	})
}

func TestDeleteFromSortedSet(t *testing.T) {

	dummyCampaignUUID := uuid.New().String()
	dummyParticipantUUID := uuid.New().String()
	currentTime := time.Now()

	dummyItem := Submission{
		UUID:            uuid.New().String(),
		ParticipantUUID: dummyParticipantUUID,
		CampaignUUID:    dummyCampaignUUID,
		Value:           "dummy submission",
		CreatedAt:       currentTime,
	}

	expectedKey := fmt.Sprintf(SORTED_SET_PARTICIPANT_SUBMISSION, dummyCampaignUUID, dummyParticipantUUID)

	t.Run("successful delete item from sorted set", func(t *testing.T) {

		redisClient, mockRedis := redismock.NewClientMock()
		mockRedis.ExpectZRem(expectedKey, dummyItem.UUID).SetVal(1)

		paginate := NewLinkedPagination[Submission](
			logger,
			redisClient,
			nil,
		)

		errorRemoveMemberFromSortedSet := paginate.RemoveItem(
			SORTED_SET_PARTICIPANT_SUBMISSION,
			dummyItem.UUID,
			&dummyItem,
			"testdeletefromsortedset",
			LogErrorSubmission,
			dummyCampaignUUID,
			dummyParticipantUUID,
		)

		assert.Nil(t, errorRemoveMemberFromSortedSet)
	})

	t.Run("remove from sorted set fatal error", func(t *testing.T) {

		redisClient, mockRedis := redismock.NewClientMock()
		mockRedis.ExpectZRem(expectedKey, dummyItem.UUID).SetErr(errors.New("fatal error: connection lost"))

		paginate := NewLinkedPagination[Submission](
			logger,
			redisClient,
			nil,
		)

		errorRemoveMemberFromSortedSet := paginate.RemoveItem(
			SORTED_SET_PARTICIPANT_SUBMISSION,
			dummyItem.UUID,
			&dummyItem,
			"testdeletefromsortedset",
			LogErrorSubmission,
			dummyCampaignUUID,
			dummyParticipantUUID,
		)

		assert.NotNil(t, errorRemoveMemberFromSortedSet)
		assert.Equal(t, REDIS_FATAL_ERROR, errorRemoveMemberFromSortedSet.Err)
		assert.Equal(t, "testdeletefromsortedset.delete_from_sorted_set_REDIS_FATAL_ERROR", errorRemoveMemberFromSortedSet.Context)
	})
}

func TestGetTotalItemSortedSet(t *testing.T) {

	dummyCampaignUUID := uuid.New().String()
	dummyParticipantUUID := uuid.New().String()

	expectedKey := fmt.Sprintf(SORTED_SET_PARTICIPANT_SUBMISSION, dummyCampaignUUID, dummyParticipantUUID)

	t.Run("success get total item", func(t *testing.T) {

		redisClient, mockRedis := redismock.NewClientMock()
		mockRedis.ExpectZCard(expectedKey).SetVal(10)

		paginate := NewLinkedPagination[Submission](
			logger,
			redisClient,
			nil,
		)

		errorGetTotalItem := paginate.TotalItem(
			SORTED_SET_PARTICIPANT_SUBMISSION,
			"testgettotal",
			dummyCampaignUUID,
			dummyParticipantUUID,
		)

		assert.Nil(t, errorGetTotalItem)
	})

	t.Run("fatal error", func(t *testing.T) {

		redisClient, mockRedis := redismock.NewClientMock()
		mockRedis.ExpectZCard(expectedKey).SetErr(errors.New("fatal error: connection lost"))

		paginate := NewLinkedPagination[Submission](
			logger,
			redisClient,
			nil,
		)

		errorGetTotalItem := paginate.TotalItem(
			SORTED_SET_PARTICIPANT_SUBMISSION,
			"testgettotal",
			dummyCampaignUUID,
			dummyParticipantUUID,
		)

		assert.NotNil(t, errorGetTotalItem)
		assert.Equal(t, REDIS_FATAL_ERROR, errorGetTotalItem.Err)
		assert.Equal(t, "testgettotal.get_total_item_sorted_set_REDIS_FATAL_ERROR", errorGetTotalItem.Context)
	})
}

func TestPaginate(t *testing.T) {

	t.Run("paginate successful", func(t *testing.T) {

	})

	t.Run("no item found", func(t *testing.T) {

	})

	t.Run("sorted set exists, lastMembers latest member exists", func(t *testing.T) {

	})

	t.Run("latest member of lastMembers null", func(t *testing.T) {

	})

	t.Run("many member of lastMembers null", func(t *testing.T) {

	})

	t.Run("all member of lastMembers null", func(t *testing.T) {

	})

	t.Run("lastMembers exceeding MAX_LAST_MEMBERS", func(t *testing.T) {

	})

	t.Run("zrevrank fatal error", func(t *testing.T) {

	})

	t.Run("zrevrange fatal error", func(t *testing.T) {

	})
}
