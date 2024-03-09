package database

import (
	"ScalableBackend/internal/entity"
	"context"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

func (g *GormDatabase) CreateTag(ctx context.Context, tag *entity.Tag) error {
	return g.queryMetric.Do("create_tag", func() error {
		err := g.db.WithContext(ctx).Create(tag).Error
		if err != nil {
			logrus.WithError(err).Error("error while creating a author")
			return err
		}
		return nil
	})
}

func (g *GormDatabase) ListTags(ctx context.Context) ([]entity.Tag, error) {
	var tags []entity.Tag
	return tags, g.queryMetric.Do("list_tags", func() error {
		err := g.db.WithContext(ctx).Find(&tags).Error
		if err != nil {
			return err
		}
		return nil
	})
}

func (g *GormDatabase) ListTagArticles(ctx context.Context, slug string) ([]entity.Article, error) {
	return g.listTagArticlesLegacy(ctx, slug)
}

type articleTagResults struct {
	Article    entity.Article `gorm:"embedded"`
	ConcatTags string
}

func (g *GormDatabase) listTagArticlesAggregate(ctx context.Context, slug string) ([]entity.Article, error) {
	var results []articleTagResults

	err := g.db.WithContext(ctx).Raw(`
SELECT articles.id, articles.created_at, articles.updated_at, articles.deleted_at, articles.title, articles.content, articles.author_id, 
       authors.id AS Author__id, authors.created_at AS Author__created_at, authors.updated_at AS Author__updated_at, authors.deleted_at AS Author__deleted_at, authors.display_name AS Author__display_name,
       GROUP_CONCAT(CONCAT(tags.slug,'/',tags.id,'/',tags.name) SEPARATOR ',') AS ConcatTags
FROM articles
INNER JOIN authors ON articles.author_id = authors.id
INNER JOIN article_tags ON articles.id = article_tags.article_id
INNER JOIN tags ON article_tags.tag_id = tags.id
WHERE articles.id IN (SELECT articles.id FROM articles INNER JOIN article_tags ON articles.id = article_tags.article_id INNER JOIN tags ON article_tags.tag_id = tags.id WHERE tags.slug = ?)
GROUP BY articles.id, articles.title, articles.content, articles.author_id, Author__id, Author__created_at, Author__updated_at, Author__deleted_at, Author__display_name
`, slug).Scan(&results).Error

	if err != nil {
		logrus.WithError(err).Error("couldn't get aggregate tag articles")
		return nil, err
	}

	for i, result := range results {
		tagPairs := strings.Split(result.ConcatTags, ",")
		tags := make([]entity.Tag, 0, len(tagPairs))
		for _, pair := range tagPairs {
			parts := strings.Split(pair, "/")
			if len(parts) != 3 {
				continue
			}
			tagID, _ := strconv.ParseUint(parts[1], 10, 32)
			tags = append(tags, entity.Tag{
				Model: gorm.Model{ID: uint(tagID)},
				Slug:  parts[0],
				Name:  parts[2],
			})
		}
		results[i].Article.Tags = tags
	}

	return lo.Map(results, func(item articleTagResults, _ int) entity.Article {
		return item.Article
	}), nil
}

func (g *GormDatabase) listTagArticlesLegacy(ctx context.Context, slug string) ([]entity.Article, error) {
	var articles []entity.Article
	return articles, g.queryMetric.Do("list_tag_articles", func() error {
		err := g.db.WithContext(ctx).
			Model(&entity.Article{}).
			Preload("Tags").
			Joins("JOIN article_tags ON article_tags.article_id = articles.id").
			Joins("JOIN tags ON tags.id = article_tags.tag_id").
			Joins("Author").
			Where("tags.slug = ?", slug).
			Find(&articles).Error
		if err != nil {
			logrus.WithError(err).WithField("slug", slug).Error("couldn't load tag articles")
			return err
		}

		return nil
	})
}
