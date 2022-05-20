package bullsbearsmon

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest"
)

type OrderPlacer interface {
	WaitForOrderExecution(
		ctx context.Context,
		accountID tinkoffinvest.AccountID,
		orderID tinkoffinvest.OrderID,
	) (*tinkoffinvest.Quotation, error)

	PlaceMarketSellOrder(ctx context.Context, request tinkoffinvest.PlaceOrderRequest) (tinkoffinvest.OrderID, error)
	PlaceMarketBuyOrder(ctx context.Context, request tinkoffinvest.PlaceOrderRequest) (tinkoffinvest.OrderID, error)
	PlaceLimitSellOrder(ctx context.Context, request tinkoffinvest.PlaceOrderRequest) (tinkoffinvest.OrderID, error)
	PlaceLimitBuyOrder(ctx context.Context, request tinkoffinvest.PlaceOrderRequest) (tinkoffinvest.OrderID, error)
}

// Strategy realize the next strategy:
// if there are more lots in buy orders than in sell orders in ToolConfig.DominanceRatio times,
// then the robot buys the instrument at the market price, otherwise it sells,
// immediately placing an order in the opposite direction, but with a certain percentage of profit.
type Strategy struct {
	account            tinkoffinvest.AccountID
	ignoreInconsistent bool
	toolConfigs        map[string]ToolConfig // by FIGI.
	orderPlacer        OrderPlacer
}

type ToolConfig struct {
	FIGI             string
	DominanceRatio   float64
	ProfitPercentage float64
}

func New(account string, ignoreInconsistent bool, tools []ToolConfig, orderPlacer OrderPlacer) (*Strategy, error) {
	confs := make(map[string]ToolConfig, len(tools))
	for _, t := range tools {
		if _, ok := confs[t.FIGI]; ok {
			return nil, fmt.Errorf("duplicated tool: %s", t.FIGI)
		}
		confs[t.FIGI] = t
	}

	return &Strategy{
		account:            tinkoffinvest.AccountID(account),
		ignoreInconsistent: ignoreInconsistent,
		toolConfigs:        confs,
		orderPlacer:        orderPlacer,
	}, nil
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

	buys := tinkoffinvest.CountLots(change.Bids)  // Bulls.
	sells := tinkoffinvest.CountLots(change.Acks) // Bears.

	buysToSells := float64(buys) / float64(sells)
	sellsToBuys := 1. / buysToSells

	logger := log.With().
		Str("strategy", s.Name()).
		Str("figi", change.FIGI).Logger()

	logger.Info().
		Int("buys", buys).
		Int("sells", sells).
		Float64("buys_to_sells", buysToSells).
		Float64("sells_to_buys", sellsToBuys).
		Msg("order book change")

	if buysToSells >= conf.DominanceRatio {
		return s.placeBuySellPair(ctx, logger, change.FIGI, conf.ProfitPercentage, change.LimitUp)
	}

	if sellsToBuys >= conf.DominanceRatio {
		return s.placeSellBuyPair(ctx, logger, change.FIGI, conf.ProfitPercentage, change.LimitDown)
	}

	return nil
}

func (s *Strategy) placeBuySellPair( //nolint:dupl
	ctx context.Context,
	logger zerolog.Logger,
	figi string,
	profit float64,
	limitUp tinkoffinvest.Quotation,
) error {
	orderID, err := s.orderPlacer.PlaceMarketBuyOrder(ctx, tinkoffinvest.PlaceOrderRequest{
		AccountID: s.account,
		FIGI:      figi,
		Lots:      1.,
	})
	if err != nil {
		return fmt.Errorf("place market buy order: %v", err)
	}

	price, err := s.orderPlacer.WaitForOrderExecution(ctx, s.account, orderID)
	if err != nil {
		return fmt.Errorf("wait for market order %s execution: %v", orderID, err)
	}

	logger.Info().
		Str("price", price.String()).
		Str("order_id", string(orderID)).
		Msg("buy tool by market")

	p := price.Mul(1. + profit)
	if p.Greater(limitUp) {
		return nil
	}

	orderID, err = s.orderPlacer.PlaceLimitSellOrder(ctx, tinkoffinvest.PlaceOrderRequest{
		AccountID: s.account,
		FIGI:      figi,
		Lots:      1,
		Price:     &p,
	})
	if err != nil {
		return fmt.Errorf("place limit sell order: %v", err)
	}

	logger.Info().
		Str("price", p.String()).
		Str("order_id", string(orderID)).
		Msg("place limit sell order")

	return nil
}

func (s *Strategy) placeSellBuyPair( //nolint:dupl
	ctx context.Context,
	logger zerolog.Logger,
	figi string,
	profit float64,
	limitDown tinkoffinvest.Quotation,
) error {
	orderID, err := s.orderPlacer.PlaceMarketSellOrder(ctx, tinkoffinvest.PlaceOrderRequest{
		AccountID: s.account,
		FIGI:      figi,
		Lots:      1.,
	})
	if err != nil {
		return fmt.Errorf("place market sell order: %v", err)
	}

	price, err := s.orderPlacer.WaitForOrderExecution(ctx, s.account, orderID)
	if err != nil {
		return fmt.Errorf("wait for market order %s execution: %v", orderID, err)
	}

	logger.Info().
		Str("price", price.String()).
		Str("order_id", string(orderID)).
		Msg("sell tool by market")

	p := price.Mul(1. - profit)
	if p.Less(limitDown) {
		return nil
	}

	orderID, err = s.orderPlacer.PlaceLimitBuyOrder(ctx, tinkoffinvest.PlaceOrderRequest{
		AccountID: s.account,
		FIGI:      figi,
		Lots:      1,
		Price:     &p,
	})
	if err != nil {
		return fmt.Errorf("place limit buy order: %v", err)
	}

	logger.Info().
		Str("price", p.String()).
		Str("order_id", string(orderID)).
		Msg("place limit buy order")

	return nil
}
