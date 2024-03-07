package controller

import (
	"ScalableBackend/internal/database"
	"ScalableBackend/internal/promhelper"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"reflect"
)

type EchoController struct {
	e              *echo.Echo
	db             database.Database
	endpointMetric *promhelper.HistogramWithCounter
}

func NewEchoController(e *echo.Echo, db database.Database) *EchoController {
	c := &EchoController{
		e:              e,
		db:             db,
		endpointMetric: promhelper.NewHistogramWithCounter("app_endpoint_response", prometheus.DefBuckets),
	}
	c.urls()
	return c
}

func (ec *EchoController) urls() {
	ec.authorUrls()
	ec.articleUrls()
	ec.tagUrls()
}

// Bind todo: no validation yet, but in the future can be added here
func Bind[T any](c echo.Context) (T, error) {
	var t T
	err := c.Bind(&t)
	if err != nil {
		logrus.WithError(err).WithField("type", reflect.TypeOf(t)).Errorln("couldn't bind the request")
	}
	return t, err
}
