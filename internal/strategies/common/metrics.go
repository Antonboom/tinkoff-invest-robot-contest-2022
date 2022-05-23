package common

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest"
)

const subsystem = "orders"

type OrderType string

const (
	OrderTypeMarketSell OrderType = "market_sell"
	OrderTypeMarketBuy  OrderType = "market_buy"
	OrderTypeLimitSell  OrderType = "limit_sell"
	OrderTypeLimitBuy   OrderType = "limit_buy"
)

var (
	ordersTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "trading_robot",
		Subsystem: subsystem,
		Name:      "total",
		Help:      "Total amount of orders",
	}, []string{"strategy", "figi", "order_type"})

	orderPrice = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "trading_robot",
		Subsystem: subsystem,
		Name:      "order_price",
		Help:      "Statistic of orders prices (requested or executed)",
	}, []string{"strategy", "figi", "order_type"})
)

func CollectOrderPrice(price float64, strategy string, figi tinkoffinvest.FIGI, orderType OrderType) {
	l := prometheus.Labels{
		"strategy":   strategy,
		"figi":       figi.S(),
		"order_type": string(orderType),
	}

	ordersTotal.With(l).Inc()
	orderPrice.With(l).Set(price)
}
