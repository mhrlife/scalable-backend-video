package cache

import (
	"ScalableBackend/internal/entity"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/allegro/bigcache/v3"
	"github.com/sirupsen/logrus"
	"time"
)

var _ Cache = &InMemoryBigCache{}

type InMemoryBigCache struct {
	client     *bigcache.BigCache
	redisCache *RedisCache
}

func NewInMemoryBigCache(redisCache *RedisCache) *InMemoryBigCache {
	bc, _ := bigcache.New(context.Background(), bigcache.DefaultConfig(10*time.Minute))
	return &InMemoryBigCache{client: bc, redisCache: redisCache}
}

func (i *InMemoryBigCache) TagArticles(ctx context.Context, tagSlug string) ([]entity.Article, error) {
	articles := make([]entity.Article, 0)
	cacheKey := fmt.Sprintf("tag_articles:" + tagSlug)
	b, err := i.client.Get(cacheKey)
	// data exists in the cache
	if err == nil {
		if err := json.Unmarshal(b, &articles); err != nil {
			logrus.WithError(err).Error("couldn't unmarshal bigcache data")
			return nil, err
		}
		return articles, nil
	}

	if !errors.Is(err, bigcache.ErrEntryNotFound) {
		logrus.WithError(err).Error("error while fetching tag articles from bigcache")
		return nil, err
	}

	// everything is fine but data is not presented in the BigCache
	articles, err = i.redisCache.TagArticles(ctx, tagSlug)
	if err != nil {
		return nil, err
	}

	b, err = json.Marshal(articles)
	if err != nil {
		logrus.WithError(err).Error("couldn't marshal articles list")
		return nil, err
	}

	return articles, i.client.Set(cacheKey, b)

}
