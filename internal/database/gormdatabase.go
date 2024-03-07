package database

import (
	"ScalableBackend/internal/entity"
	"ScalableBackend/internal/promhelper"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var _ Database = &GormDatabase{}

type GormDatabase struct {
	db          *gorm.DB
	queryMetric *promhelper.HistogramWithCounter
}

func NewGormDatabase(db *gorm.DB) *GormDatabase {
	return &GormDatabase{
		db:          db,
		queryMetric: promhelper.NewHistogramWithCounter("app_database_queries", prometheus.DefBuckets),
	}
}

func (g *GormDatabase) Migrate() error {
	err := g.db.AutoMigrate(
		&entity.Author{},
		&entity.Tag{},
		&entity.Article{},
	)
	if err != nil {
		logrus.WithError(err).Error("error while auto migrating the database with gorm")
		return err
	}
	return nil
}
