package controller

import (
	"ScalableBackend/internal/entity"
	"context"
	"github.com/labstack/echo/v4"
	"time"
)

func (ec *EchoController) articleUrls() {
	g := ec.e.Group("/article")
	g.POST("/", ec.createArticle)
	g.GET("/", ec.listArticles)
}

type createArticleRequest struct {
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	AuthorID uint     `json:"author_id"`
	TagSlugs []string `json:"tag_slugs"`
}

func (ec *EchoController) createArticle(c echo.Context) error {
	return ec.endpointMetric.Do("create_tag", func() error {
		request, err := Bind[createArticleRequest](c)
		if err != nil {
			return err
		}

		// write requests can not be canceled by the client
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		article := entity.Article{
			Title:    request.Title,
			Content:  request.Content,
			AuthorID: request.AuthorID,
		}
		if err := ec.db.CreateArticle(ctx, &article, request.TagSlugs); err != nil {
			_ = c.String(500, err.Error())
			return err
		}

		return c.JSON(201, article)
	})
}

func (ec *EchoController) listArticles(c echo.Context) error {
	return ec.endpointMetric.Do("list_articles", func() error {
		articles, err := ec.db.ListArticles(c.Request().Context())
		if err != nil {
			return err
		}
		return c.JSON(200, articles)
	})
}
