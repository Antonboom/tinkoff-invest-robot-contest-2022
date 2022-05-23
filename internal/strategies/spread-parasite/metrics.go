package spreadparasite

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	subsystem = "spread_parasite"

	bestPriceTypeToSell = "to_sell"
	bestPriceTypeToBuy  = "to_buy"
)

var bestPriceGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "trading_robot",
	Subsystem: subsystem,
	Name:      "best_price",
	Help:      "Spread statistic",
}, []string{"best_price_type", "figi"})
