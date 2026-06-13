package metrics

import "github.com/prometheus/client_golang/prometheus"

var TelegramRequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "telegram_requests_total",
		Help: "Total Telegram requests by type",
	},
	[]string{"type"},
)

var AIRequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "ai_requests_total",
		Help: "Total AI requests by provider and type",
	},
	[]string{"provider", "type"},
)

var AIErrorsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "ai_errors_total",
		Help: "Total AI errors by provider and type",
	},
	[]string{"provider", "type"},
)

func Register() {
	prometheus.MustRegister(
		TelegramRequestsTotal,
		AIRequestsTotal,
		AIErrorsTotal,
	)
}
