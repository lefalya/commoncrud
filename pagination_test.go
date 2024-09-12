package commonpagination

import (
	"log/slog"

	"github.com/lefalya/commonpagination/interfaces"
	"github.com/redis/go-redis/v9"
)

func initTestPaginationType[T interfaces.Item](
	pagKeyFormat string,
	itemKeyFormat string,
	logger *slog.Logger,
	redisClient *redis.Client,
	itemCache interfaces.ItemCache[T],
) *PaginationType[T] {
	return &PaginationType[T]{
		pagKeyFormat:  pagKeyFormat,
		itemKeyFormat: itemKeyFormat,
		logger:        logger,
		redisClient:   redisClient,
		itemCache:     itemCache,
	}
}
