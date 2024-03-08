package main

import (
	"ScalableBackend/internal/controller"
	"ScalableBackend/internal/database"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	_ "go.uber.org/automaxprocs"
	"net/http"
	"os"
	"strings"
)

func main() {
	// loading .env file for the local development
	_ = godotenv.Load()
	// setup mysql connections
	masterDSN := os.Getenv("MASTER_DSN")
	replicaDSNs := strings.Split(os.Getenv("REPLICA_DSN"), ",")
	replicaDSNs = lo.Map(replicaDSNs, func(item string, _ int) string {
		return strings.TrimSpace(item)
	})
	db, err := database.NewGorm(strings.TrimSpace(masterDSN), replicaDSNs...)
	if err != nil {
		logrus.WithError(err).Panicln("the database connection setup failed")
	}
	// set up the database
	gdb := database.NewGormDatabase(db)
	if err := gdb.Migrate(); err != nil {
		logrus.WithError(err).Panicln("error while migrating the database")
	}

	// set up Prometheus exposer
	http.Handle("/metrics", promhttp.Handler())
	logrus.Info("starting the metric server on port 8081")
	go func() {
		err := http.ListenAndServe(":8081", nil)
		if err != nil {
			logrus.WithError(err).Error("the metric server has stopped")
		}
	}()
	// set up the http apis
	e := echo.New()
	controller.NewEchoController(e, gdb)
	logrus.Info("starting the api server on port 8080")
	logrus.WithError(err).Error(e.Start(":8080"))
}
