package cache

import (
	"ScalableBackend/internal/entity"
	"context"
)

type Cache interface {
	TagArticles(ctx context.Context, tagSlug string) ([]entity.Article, error)
}
