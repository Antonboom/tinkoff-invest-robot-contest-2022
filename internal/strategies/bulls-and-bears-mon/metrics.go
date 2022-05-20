package bullsbearsmon

import "github.com/prometheus/client_golang/prometheus"
import "github.com/prometheus/client_golang/prometheus/promauto"

const (
	subsystem = "bbmon"

	ratioTypeBuyToSells  = "buys_to_sells"
	ratioTypeSellsToBuys = "sells_to_buys"
)

var (
	configuredDominanceRatio = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "trading_robot",
		Subsystem: subsystem,
		Name:      "dominance_ratio",
	}, []string{"figi"})

	tradersRatioGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "trading_robot",
		Subsystem: subsystem,
		Name:      "traders_ratio",
	}, []string{"type", "figi"})
)
