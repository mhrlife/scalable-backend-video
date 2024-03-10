package entity

import (
	"encoding/json"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type Author struct {
	gorm.Model

	DisplayName string `json:"display_name"`
}

func (a Author) RedisJson() string {
	m := map[string]any{
		"id":           a.ID,
		"display_name": a.DisplayName,
	}
	b, _ := json.Marshal(m)
	return string(b)
}

type Tag struct {
	gorm.Model

	Slug string `json:"slug" gorm:"index"`
	Name string `json:"name"`
}

func (t Tag) RedisJson() string {
	m := map[string]any{
		"name": t.Name,
		"slug": t.Slug,
		"id":   t.ID,
	}
	b, _ := json.Marshal(m)
	return string(b)
}

type Article struct {
	gorm.Model

	Title   string `json:"title"`
	Content string `json:"content"`

	AuthorID uint   `json:"author_id" gorm:"index"`
	Author   Author `json:"author"`

	Tags []Tag `gorm:"many2many:article_tags;" json:"tags"`
}

func (a Article) RedisJson() string {
	m := map[string]any{
		"title":     a.Title,
		"content":   a.Content,
		"author_id": a.AuthorID,
		"tags": lo.Map(a.Tags, func(item Tag, _ int) string {
			return item.Slug
		}),
		"id": a.ID,
	}
	b, _ := json.Marshal(m)
	return string(b)
}
