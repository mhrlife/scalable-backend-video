package database

import (
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/dbresolver"
	"time"
)

func NewGorm(masterDSN string, replicaDSNs ...string) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(masterDSN), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Silent),
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})

	if err != nil {
		logrus.WithError(err).WithField("dsn", masterDSN).Error("couldn't connect to the database")
		return nil, err
	}

	if err := db.Use(dbresolver.Register(dbresolver.Config{
		Replicas: lo.Map(append(replicaDSNs, masterDSN), func(item string, _ int) gorm.Dialector {
			return mysql.Open(item)
		}),
	})); err != nil {
		logrus.WithError(err).Error("couldn't setup replica databases")
		return nil, err
	}

	sqlDB, _ := db.DB()

	sqlDB.SetMaxIdleConns(100)
	sqlDB.SetMaxOpenConns(450)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}
