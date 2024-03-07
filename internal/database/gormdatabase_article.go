package database

import (
	"ScalableBackend/internal/entity"
	"context"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func (g *GormDatabase) CreateArticle(ctx context.Context, article *entity.Article, tagSlugs []string) error {
	return g.queryMetric.Do("create_article", func() error {
		return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			err := tx.Create(article).Error
			if err != nil {
				logrus.WithError(err).Error("error while creating a article")
				return err
			}

			var tags []entity.Tag
			if err := tx.Where("slug IN ?", tagSlugs).Find(&tags).Error; err != nil {
				logrus.WithError(err).Error("error while fetching tags for article insertion")
				return err
			}

			if err := tx.Model(&article).Association("Tags").Append(tags); err != nil {
				logrus.WithError(err).Error("error while updating tags for article insertion")
				return err
			}

			return nil
		})

	})
}

func (g *GormDatabase) ListArticles(ctx context.Context) ([]entity.Article, error) {
	var articles []entity.Article
	return articles, g.queryMetric.Do("list_articles", func() error {
		err := g.db.WithContext(ctx).InnerJoins("Author").Preload("Tags").Find(&articles).Error
		if err != nil {
			logrus.WithError(err).Error("couldn't list articles")
			return err
		}
		return nil
	})
}
