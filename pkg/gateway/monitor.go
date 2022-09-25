package gateway

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	TenantHTTPRequestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tenant_http_request_counter",
			Help: "租户请求总数",
		}, []string{"tenant"},
	)
	httpRequestCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "gateway_http_request_counter",
		Help: "网关请求总数",
	})
	UserRequestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "user_http_request_counter",
		Help: "用户请求总数",
	}, []string{"user"})
)

func (g *Gateway) initMonitor() {

	prometheus.MustRegister(TenantHTTPRequestCounter, httpRequestCounter, UserRequestCounter)
	// Prometheus
	http.Handle("/metrics", promhttp.Handler())
}
