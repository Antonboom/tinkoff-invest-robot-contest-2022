package bullsbearsmon

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	subsystem = "bbmon"

	ratioTypeBuyToSells  = "buys_to_sells"
	ratioTypeSellsToBuys = "sells_to_buys"

	lotsTypeForSell = "for_sell"
	lotsTypeForBuy  = "for_buy"
)

var (
	configuredDominanceRatio = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "trading_robot",
		Subsystem: subsystem,
		Name:      "dominance_ratio",
		Help:      "The configured min ratio of sell orders to buy orders (and visa versa)",
	}, []string{"figi"})

	ordersRatio = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "trading_robot",
		Subsystem: subsystem,
		Name:      "orders_ratio",
		Help:      "The current ratio of sell orders to buy orders (and visa versa)",
	}, []string{"ratio_type", "figi"})

	tradedLots = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "trading_robot",
		Subsystem: subsystem,
		Name:      "traded_lots",
		Help:      "The current amount of orders (in lots)",
	}, []string{"lots_type", "figi"})
)
