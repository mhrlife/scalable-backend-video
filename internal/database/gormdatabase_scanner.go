package database

import (
	"ScalableBackend/internal/entity"
	"context"
	"github.com/sirupsen/logrus"
)

func (g *GormDatabase) Scan(ctx context.Context, lastId uint, perPage int) ([]entity.Article, uint, error) {
	articles := make([]entity.Article, 0)
	if err := g.db.WithContext(ctx).Model(&entity.Article{}).Preload("Tags").Joins("Author").
		Where("articles.id > ?", lastId).Order("articles.id ASC").
		Limit(perPage).Find(&articles).Error; err != nil {
		logrus.WithError(err).Errorln("error while scanning articles")
		return nil, 0, err
	}

	if len(articles) == 0 {
		return articles, 0, nil
	}

	return articles, articles[len(articles)-1].ID, nil
}
