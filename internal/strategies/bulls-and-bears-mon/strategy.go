package bullsbearsmon

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest"
	toolscache "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/services/tools-cache"
	"github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/strategies/common"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/strategy_generated.go -package bullsbearsmonmocks OrderPlacer,ToolsCache

const (
	applyingTimeout = 3 * time.Second
	lotsInTrade     = 1
)

type l = prometheus.Labels

type OrderPlacer interface {
	SubscribeForOrderBookChanges(ctx context.Context, reqs []tinkoffinvest.OrderBookRequest) (<-chan tinkoffinvest.OrderBookChange, error) //nolint:lll
	WaitForOrderExecution(ctx context.Context, _ tinkoffinvest.AccountID, _ tinkoffinvest.OrderID) (decimal.Decimal, error)

	PlaceMarketSellOrder(ctx context.Context, request tinkoffinvest.PlaceOrderRequest) (tinkoffinvest.OrderID, error)
	PlaceMarketBuyOrder(ctx context.Context, request tinkoffinvest.PlaceOrderRequest) (tinkoffinvest.OrderID, error)
	PlaceLimitSellOrder(ctx context.Context, request tinkoffinvest.PlaceOrderRequest) (tinkoffinvest.OrderID, error)
	PlaceLimitBuyOrder(ctx context.Context, request tinkoffinvest.PlaceOrderRequest) (tinkoffinvest.OrderID, error)
}

type ToolsCache interface {
	Get(ctx context.Context, figi tinkoffinvest.FIGI) (toolscache.Tool, error)
}

// Strategy realize the next strategy:
// if there are more lots in buy orders than in sell orders in ToolConfig.DominanceRatio times,
// then the robot buys the instrument at the market price, otherwise it sells,
// immediately placing an order in the opposite direction, but with a certain percentage of profit.
type Strategy struct {
	account            tinkoffinvest.AccountID
	ignoreInconsistent bool
	toolConfigs        map[tinkoffinvest.FIGI]ToolConfig

	orderPlacer OrderPlacer
	toolsCache  ToolsCache
	logger      zerolog.Logger
}

type ToolConfig struct {
	FIGI             tinkoffinvest.FIGI
	Depth            int
	DominanceRatio   float64
	ProfitPercentage float64

	// stocksPerLot fetched automatically at the start.
	stocksPerLot int
	// minPriceInc fetched automatically at the start.
	minPriceInc decimal.Decimal
}

func New(
	account tinkoffinvest.AccountID,
	ignoreInconsistent bool,
	tools []ToolConfig,
	orderPlacer OrderPlacer,
	toolsCache ToolsCache,
) (*Strategy, error) {
	confs := make(map[tinkoffinvest.FIGI]ToolConfig, len(tools))
	for _, t := range tools {
		if _, ok := confs[t.FIGI]; ok {
			return nil, fmt.Errorf("duplicated tool: %s", t.FIGI)
		}

		confs[t.FIGI] = t
		configuredDominanceRatio.With(l{"figi": t.FIGI.S()}).Set(t.DominanceRatio)
	}

	s := &Strategy{
		account:            account,
		ignoreInconsistent: ignoreInconsistent,
		toolConfigs:        confs,
		orderPlacer:        orderPlacer,
		toolsCache:         toolsCache,
	}
	s.logger = log.With().Str("strategy", s.Name()).Logger()

	return s, nil
}

func (s *Strategy) Name() string {
	return "bulls-and-bears-monitoring"
}

// Run starts order book monitoring and calls Apply on every new change.
func (s *Strategy) Run(ctx context.Context) error {
	if err := s.fetchToolConfigs(ctx); err != nil {
		return fmt.Errorf("fetch tool configs: %v", err)
	}

	reqs := make([]tinkoffinvest.OrderBookRequest, 0, len(s.toolConfigs))
	for _, t := range s.toolConfigs {
		reqs = append(reqs, tinkoffinvest.OrderBookRequest{
			FIGI:  t.FIGI,
			Depth: t.Depth,
		})
	}

	changes, err := s.orderPlacer.SubscribeForOrderBookChanges(ctx, reqs)
	if err != nil {
		return fmt.Errorf("subscribe for order book changes: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-time.After(5 * time.Second):
			s.logger.Debug().Msg("no order book changes due to period")

		case change, ok := <-changes:
			if !ok {
				return nil
			}

			func() {
				ctx, cancel := context.WithTimeout(ctx, applyingTimeout)
				defer cancel()

				if err := s.Apply(ctx, change); err != nil {
					s.logger.Err(err).Msg("cannot apply order book change")
				}
			}()
		}
	}
}

func (s *Strategy) fetchToolConfigs(ctx context.Context) error {
	s.logger.Debug().Msg("fetch tool info")

	for i, t := range s.toolConfigs {
		tool, err := s.toolsCache.Get(ctx, t.FIGI)
		if err != nil {
			return fmt.Errorf("get cached tool %v: %v", t.FIGI, err)
		}

		t.stocksPerLot = tool.StocksPerLot
		t.minPriceInc = tool.MinPriceInc
		s.toolConfigs[i] = t
	}
	return nil
}

// Apply applies Strategy to the next order book change.
func (s *Strategy) Apply(ctx context.Context, change tinkoffinvest.OrderBookChange) error {
	logger := s.logger.With().Str("figi", change.FIGI.S()).Logger()

	if s.ignoreInconsistent && change.IsConsistent {
		logger.Debug().Msg("ignore inconsistent order book change")
		return nil
	}

	conf, ok := s.toolConfigs[change.FIGI]
	if !ok {
		return fmt.Errorf("not found config for tool %q", change.FIGI)
	}

	buys := tinkoffinvest.CountLots(change.Bids)  // Bulls.
	sells := tinkoffinvest.CountLots(change.Asks) // Bears.

	tradedLots.With(l{"lots_type": lotsTypeForBuy, "figi": change.FIGI.S()}).Set(float64(buys))
	tradedLots.With(l{"lots_type": lotsTypeForSell, "figi": change.FIGI.S()}).Set(float64(sells))

	buysToSells := float64(buys) / float64(sells)
	sellsToBuys := 1. / buysToSells

	ordersRatio.With(l{"ratio_type": ratioTypeBuyToSells, "figi": change.FIGI.S()}).Set(buysToSells)
	ordersRatio.With(l{"ratio_type": ratioTypeSellsToBuys, "figi": change.FIGI.S()}).Set(sellsToBuys)

	logger.Info().
		Int("buys", buys).
		Int("sells", sells).
		Float64("buys_to_sells", buysToSells).
		Float64("sells_to_buys", sellsToBuys).
		Msg("order book change")

	if buysToSells >= conf.DominanceRatio {
		return s.placeBuySellPair(ctx, logger, conf, change.LimitUp)
	}

	if sellsToBuys >= conf.DominanceRatio {
		return s.placeSellBuyPair(ctx, logger, conf, change.LimitDown)
	}

	return nil
}

func (s *Strategy) placeBuySellPair(
	ctx context.Context,
	logger zerolog.Logger,
	conf ToolConfig,
	limitUp decimal.Decimal,
) error {
	orderID, err := s.orderPlacer.PlaceMarketBuyOrder(ctx, tinkoffinvest.PlaceOrderRequest{
		AccountID: s.account,
		FIGI:      conf.FIGI,
		Lots:      lotsInTrade,
	})
	if err != nil {
		if errors.Is(err, tinkoffinvest.ErrNotEnoughStocks) {
			return nil
		}
		return fmt.Errorf("place market buy order: %v", err)
	}

	executedPrice, err := s.orderPlacer.WaitForOrderExecution(ctx, s.account, orderID)
	if err != nil {
		return fmt.Errorf("wait for market order %s execution: %v", orderID, err)
	}
	p := executedPrice.Div(decimal.NewFromInt(int64(conf.stocksPerLot) * lotsInTrade))

	common.CollectOrderPrice(p.InexactFloat64(), s.Name(), conf.FIGI, common.OrderTypeMarketBuy)
	logger.Info().
		Str("share_price", p.String()).
		Str("order_id", orderID.S()).
		Msg("buy lots by market")

	p = common.RoundToMinPriceIncrement(
		p.Mul(decimal.NewFromFloat(1.+conf.ProfitPercentage)),
		conf.minPriceInc,
	)
	if p.GreaterThan(limitUp) {
		return nil
	}

	orderID, err = s.orderPlacer.PlaceLimitSellOrder(ctx, tinkoffinvest.PlaceOrderRequest{
		AccountID: s.account,
		FIGI:      conf.FIGI,
		Lots:      lotsInTrade,
		Price:     p,
	})
	if err != nil {
		if errors.Is(err, tinkoffinvest.ErrNotEnoughStocks) {
			return nil
		}
		return fmt.Errorf("place limit sell order: %v", err)
	}

	common.CollectOrderPrice(p.InexactFloat64(), s.Name(), conf.FIGI, common.OrderTypeLimitSell)
	logger.Info().
		Str("price", p.String()).
		Str("order_id", orderID.S()).
		Msg("place limit sell order")

	return nil
}

func (s *Strategy) placeSellBuyPair(
	ctx context.Context,
	logger zerolog.Logger,
	conf ToolConfig,
	limitDown decimal.Decimal,
) error {
	orderID, err := s.orderPlacer.PlaceMarketSellOrder(ctx, tinkoffinvest.PlaceOrderRequest{
		AccountID: s.account,
		FIGI:      conf.FIGI,
		Lots:      lotsInTrade,
	})
	if err != nil {
		if errors.Is(err, tinkoffinvest.ErrNotEnoughStocks) {
			return nil
		}
		return fmt.Errorf("place market sell order: %v", err)
	}

	executedPrice, err := s.orderPlacer.WaitForOrderExecution(ctx, s.account, orderID)
	if err != nil {
		return fmt.Errorf("wait for market order %s execution: %v", orderID, err)
	}
	p := executedPrice.Div(decimal.NewFromInt(int64(conf.stocksPerLot) * lotsInTrade))

	common.CollectOrderPrice(p.InexactFloat64(), s.Name(), conf.FIGI, common.OrderTypeMarketSell)
	logger.Info().
		Str("share_price", p.String()).
		Str("order_id", orderID.S()).
		Msg("sell lots by market")

	p = common.RoundToMinPriceIncrement(
		p.Mul(decimal.NewFromFloat(1.-conf.ProfitPercentage)),
		conf.minPriceInc,
	)
	if p.LessThan(limitDown) {
		return nil
	}

	orderID, err = s.orderPlacer.PlaceLimitBuyOrder(ctx, tinkoffinvest.PlaceOrderRequest{
		AccountID: s.account,
		FIGI:      conf.FIGI,
		Lots:      lotsInTrade,
		Price:     p,
	})
	if err != nil {
		if errors.Is(err, tinkoffinvest.ErrNotEnoughStocks) {
			return nil
		}
		return fmt.Errorf("place limit buy order: %v", err)
	}

	common.CollectOrderPrice(p.InexactFloat64(), s.Name(), conf.FIGI, common.OrderTypeLimitBuy)
	logger.Info().
		Str("price", p.String()).
		Str("order_id", orderID.S()).
		Msg("place limit buy order")

	return nil
}
