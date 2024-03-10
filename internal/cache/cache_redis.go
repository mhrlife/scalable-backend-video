package cache

import (
	"ScalableBackend/internal/entity"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/rueidis"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"strings"
)

var _ Cache = &RedisCache{}

type RedisCache struct {
	rdb rueidis.Client
}

func NewRedisCache(rdb rueidis.Client) *RedisCache {
	return &RedisCache{rdb: rdb}
}

type cachedArticle struct {
	entity.Article
	Tags []string `json:"tags"`
}

func (r RedisCache) TagArticles(ctx context.Context, tagSlug string) ([]entity.Article, error) {
	_, result, err := r.rdb.Do(ctx,
		r.rdb.B().FtSearch().Index("idx_articles").Query(fmt.Sprintf("@tags:{%s}", tagSlug)).Build(),
	).AsFtSearch()
	if err != nil {
		logrus.WithError(err).Errorln("couldn't fetch tag articles from redis")
		return nil, err
	}

	jsonMGets := make(map[string]string)
	articles := make([]cachedArticle, 0)
	for _, docJson := range result {
		var article cachedArticle
		if err := json.Unmarshal([]byte(docJson.Doc["$"]), &article); err != nil {
			logrus.WithError(err).Errorln("couldn't unmarshal the article from redis")
			return nil, err
		}
		authorId := article.AuthorID
		jsonMGets[fmt.Sprintf("author:%d", authorId)] = ""
		for _, tag := range article.Tags {
			jsonMGets[fmt.Sprintf("tag:%s", tag)] = ""
		}
		articles = append(articles, article)
	}

	keys := lo.Keys(jsonMGets)
	attributes, err := r.rdb.Do(ctx,
		r.rdb.B().JsonMget().Key(keys...).Path(".").Build(),
	).ToArray()
	if err != nil {
		logrus.WithError(err).Errorln("couldn't get post attrs")
		return nil, err
	}

	authors := make(map[uint]entity.Author)
	tags := make(map[string]entity.Tag)
	for i, key := range keys {
		val, _ := attributes[i].ToString()
		if strings.HasPrefix(key, "author:") {
			var author entity.Author
			if err := json.Unmarshal([]byte(val), &author); err != nil {
				logrus.WithError(err).Errorln("couldn't unmarshal author info")
			}
			authors[author.ID] = author
		}
		if strings.HasPrefix(key, "tag:") {
			var tag entity.Tag
			if err := json.Unmarshal([]byte(val), &tag); err != nil {
				logrus.WithError(err).Errorln("couldn't unmarshal tag info")
			}
			tags[tag.Slug] = tag
		}
	}

	// now appending the attributes

	return lo.Map(articles, func(item cachedArticle, _ int) entity.Article {
		item.Article.Tags = lo.FilterMap(item.Tags, func(tagSlug string, _ int) (entity.Tag, bool) {
			tag, ok := tags[tagSlug]
			if !ok {
				logrus.WithField("tag", tagSlug).Error("couldn't fetch tag from redis")
				return entity.Tag{}, false
			}
			return tag, true
		})

		item.Article.Author = authors[item.AuthorID]
		return item.Article
	}), nil
}
