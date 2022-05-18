package bullsbearsmon

import (
	"context"
	"fmt"
	"github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

type OrderPlacer interface {
	PlaceMarketSellOrder(ctx context.Context, request tinkoffinvest.PlaceOrderRequest) (*tinkoffinvest.PlaceOrderResponse, error)
	PlaceMarketBuyOrder(ctx context.Context, request tinkoffinvest.PlaceOrderRequest) (*tinkoffinvest.PlaceOrderResponse, error)
	PlaceLimitSellOrder(ctx context.Context, request tinkoffinvest.PlaceOrderRequest) (*tinkoffinvest.PlaceOrderResponse, error)
	PlaceLimitBuyOrder(ctx context.Context, request tinkoffinvest.PlaceOrderRequest) (*tinkoffinvest.PlaceOrderResponse, error)
}

// Strategy realize the next strategy:
// if there are more lots in buy orders than in sell orders in ToolConfig.DominanceRatio times,
// then the robot buys the instrument at the market price, otherwise it sells,
// immediately placing an order in the opposite direction, but with a certain percentage of profit.
type Strategy struct {
	account            string
	ignoreInconsistent bool
	toolConfigs        map[string]ToolConfig // by FIGI.
	orderPlacer        OrderPlacer
	logger             zerolog.Logger
}

type ToolConfig struct {
	FIGI           string
	DominanceRatio float64
}

func New(account string, ignoreInconsistent bool, tools []ToolConfig, orderPlacer OrderPlacer) (*Strategy, error) {
	confs := make(map[string]ToolConfig, len(tools))
	for _, t := range tools {
		if _, ok := confs[t.FIGI]; ok {
			return nil, fmt.Errorf("duplicated tool: %s", t.FIGI)
		}
		confs[t.FIGI] = t
	}

	s := &Strategy{
		account:            account,
		ignoreInconsistent: ignoreInconsistent,
		toolConfigs:        confs,
		orderPlacer:        orderPlacer,
	}
	s.logger = log.With().Str("strategy", s.Name()).Logger()

	return s, nil
}

func (s *Strategy) Name() string {
	return "bulls-and-bears-monitoring"
}

func (s *Strategy) Apply(ctx context.Context, change tinkoffinvest.OrderBookChange) error {
	if s.ignoreInconsistent && change.IsConsistent {
		return nil
	}

	conf, ok := s.toolConfigs[change.FIGI]
	if !ok {
		return fmt.Errorf("not found config for tool %q", change.FIGI)
	}

	buys := countLots(change.Bids)  // Bulls.
	sells := countLots(change.Acks) // Bears.

	buysToSells := float64(buys) / float64(sells)
	sellsToBuys := float64(sells) / float64(buys)

	s.logger.Info().
		Str("figi", change.FIGI).
		Int("buys", buys).
		Int("sells", sells).
		Float64("buys_to_sells", buysToSells).
		Float64("sells_to_buys", sellsToBuys).
		Msg("order book change")

	if buysToSells >= conf.DominanceRatio {
		_, err := s.orderPlacer.PlaceMarketBuyOrder(ctx, tinkoffinvest.PlaceOrderRequest{
			AccountID: s.account,
			FIGI:      change.FIGI,
			Lots:      1.,
		})
		if err != nil {
			return fmt.Errorf("place market buy order: %v", err)
		}
		os.Exit(0)

	} else if sellsToBuys >= conf.DominanceRatio {
		_, err := s.orderPlacer.PlaceMarketSellOrder(ctx, tinkoffinvest.PlaceOrderRequest{
			AccountID: s.account,
			FIGI:      change.FIGI,
			Lots:      1.,
		})
		if err != nil {
			return fmt.Errorf("place market sell order: %v", err)
		}
		os.Exit(0)
	}
	return nil
}

func countLots(orders []tinkoffinvest.Order) int {
	var result int
	for _, o := range orders {
		result += o.Lots
	}
	return result
}
