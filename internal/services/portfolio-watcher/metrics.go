package portfoliowatcher

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const subsystem = "portfolio"

var (
	currentBalance = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "trading_robot",
		Subsystem: subsystem,
		Name:      "balance",
		Help:      "Current account balance",
	}, []string{"account_number"})

	sharesTotalPrice = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "trading_robot",
		Subsystem: subsystem,
		Name:      "shares_total_price",
		Help:      "Total portfolio shares price",
	}, []string{"account_number"})

	shareQuantity = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "trading_robot",
		Subsystem: subsystem,
		Name:      "share_quantity",
		Help:      "Portfolio share quantity",
	}, []string{"account_number", "figi"})

	shareAvgPrice = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "trading_robot",
		Subsystem: subsystem,
		Name:      "share_avg_price",
		Help:      "Portfolio share average price",
	}, []string{"account_number", "figi"})
)
