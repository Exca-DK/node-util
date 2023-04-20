package metrics

import (
	"net/http"

	"github.com/Exca-DK/node-util/service/metrics"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var Factory metrics.Factory
var handler http.Handler

func init() {
	r := metrics.NewRegistry()
	Factory = metrics.With(r)
	handler = promhttp.HandlerFor(r, promhttp.HandlerOpts{})
}

func RegisterMetricsHandler(eng *gin.Engine) {
	collectors.NewBuildInfoCollector()
	eng.GET("/metrics", func(c *gin.Context) { handler.ServeHTTP(c.Writer, c.Request) })
	metrics.Enable()
}

func Enabled() bool {
	return metrics.Enabled()
}
