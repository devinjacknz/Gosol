package monitoring

import (
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    wsConnections = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "websocket_connections",
        Help: "Number of active WebSocket connections",
    })
    wsMessages = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "websocket_messages_total",
            Help: "Total number of WebSocket messages",
        },
        []string{"type", "channel"},
    )
)

func init() {
    prometheus.MustRegister(wsConnections, wsMessages)
}

func RecordMessage(msgType, channel string) {
    wsMessages.WithLabelValues(msgType, channel).Inc()
}

func Setup(r *gin.Engine) {
    r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

func IncrementWSConnections() {
    wsConnections.Inc()
}

func DecrementWSConnections() {
    wsConnections.Dec()
}
