package cache

import (
	"ScalableBackend/internal/entity"
	"context"
	"errors"
	"fmt"
	"github.com/redis/rueidis"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"strconv"
	"strings"
	"time"
)

type DBScanner interface {
	Scan(ctx context.Context, lastId uint, perPage int) ([]entity.Article, uint, error)
}

type Sync struct {
	rdb         rueidis.Client
	scanner     DBScanner
	perPageScan int
}

func NewSync(rdb rueidis.Client, scanner DBScanner) *Sync {
	s := &Sync{
		rdb:         rdb,
		scanner:     scanner,
		perPageScan: 20,
	}
	if err := s.migrate(); err != nil {
		logrus.WithError(err).Panicln("couldn't migrate redis indexes")
	}
	go s.syncCacheWorker()
	return s
}

func (s *Sync) syncCacheWorker() {
	t := time.NewTicker(time.Minute)
	for {
		if s.mustSynchronize() {
			s.syncCache()
		}
		<-t.C
	}
}

func (s *Sync) syncCache() {
	lastId := uint(0)
	for {
		articles, cursor, err := s.scanner.Scan(context.Background(), lastId, s.perPageScan)
		if err != nil {
			logrus.WithError(err).Errorln("error while scanning articles to update cache")
			return
		}

		if cursor == 0 {
			return
		}

		if err := s.updateArticles(articles); err != nil {
			return
		}

		lastId = cursor
	}

}

func (s *Sync) updateArticles(articles []entity.Article) error {
	commands := make([]rueidis.Completed, 0)
	updatedAuthors := make(map[uint]struct{})
	updatedTags := make(map[uint]struct{})
	for _, article := range articles {
		commands = append(commands,
			s.rdb.B().JsonSet().Key(fmt.Sprintf("article:%d", article.ID)).Path("$").
				Value(article.RedisJson()).Build(),
		)
		for _, tag := range article.Tags {
			if _, ok := updatedTags[tag.ID]; !ok {
				commands = append(commands,
					s.rdb.B().JsonSet().Key(fmt.Sprintf("tag:%s", tag.Slug)).Path("$").
						Value(tag.RedisJson()).Build(),
				)
				updatedTags[tag.ID] = struct{}{}
			}
		}
		if _, ok := updatedAuthors[article.AuthorID]; !ok {
			commands = append(commands,
				s.rdb.B().JsonSet().Key(fmt.Sprintf("author:%d", article.AuthorID)).Path("$").
					Value(article.Author.RedisJson()).Build(),
			)
			updatedAuthors[article.AuthorID] = struct{}{}
		}
	}
	result := s.rdb.DoMulti(context.Background(), commands...)
	err := lo.Reduce(result, func(agg error, item rueidis.RedisResult, _ int) error {
		if agg != nil {
			return agg
		}
		if item.Error() != nil {
			return item.Error()
		}
		return nil
	}, nil)
	if err != nil {
		logrus.WithError(err).Errorln("couldn't run multi sync redis")
		return err
	}
	return nil

}

func (s *Sync) mustSynchronize() bool {
	cmd := s.rdb.B().Getset().Key("config:last-sync").Value(strconv.FormatInt(time.Now().Unix(), 10)).Build()
	lastSyncTimeStr, err := s.rdb.Do(context.Background(), cmd).ToString()
	if err != nil {
		// it means no other workers have ever synchronized the cache
		if errors.Is(err, rueidis.Nil) {
			return true
		}
		logrus.WithError(err).Error("error while checking cache sync config")
		return false
	}

	timestamp, _ := strconv.ParseInt(lastSyncTimeStr, 10, 64)
	if timestamp == 0 {
		return true
	}
	// each 15 minutes the cache must be synchronized
	return time.Since(time.Unix(timestamp, 0)) > time.Minute*15
}

func (s *Sync) migrate() error {
	articleIndex := s.rdb.B().FtCreate().Index("idx_articles").OnJson().Prefix(1).Prefix("article:").
		Schema().FieldName("$.title").As("title").Text().FieldName("$.author_id").As("author_id").
		Numeric().Sortable().FieldName("$.tags.*").As("tags").Tag().Build()
	if err := s.rdb.Do(context.Background(), articleIndex).Error(); err != nil && !isIndexAlreadyExistsError(err) {
		logrus.WithError(err).Error("couldn't create the article index")
		return err
	}

	tagIndex := s.rdb.B().FtCreate().Index("idx_tag").OnJson().Prefix(1).Prefix("tag:").
		Schema().FieldName("$.id").As("id").Numeric().Sortable().Build()
	if err := s.rdb.Do(context.Background(), tagIndex).Error(); err != nil && !isIndexAlreadyExistsError(err) {
		logrus.WithError(err).Error("couldn't create the tag index")
		return err
	}

	return nil
}

func isIndexAlreadyExistsError(err error) bool {
	return strings.HasPrefix(err.Error(), "Index already exists")
}
