package spreadparasite

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

//go:generate mockgen -source=$GOFILE -destination=mocks/strategy_generated.go -package spreadparasitemocks OrderPlacer,ToolsCache

const (
	applyingTimeout = 3 * time.Second

	orderBookDepth = 1 // We are interested in border orders only.
	maxTools       = 10
	lotsInTrade    = 1
)

type l = prometheus.Labels

type OrderPlacer interface {
	SubscribeForOrderBookChanges(ctx context.Context, reqs []tinkoffinvest.OrderBookRequest) (<-chan tinkoffinvest.OrderBookChange, error) //nolint:lll
	GetTradeAvailableShares(ctx context.Context) ([]tinkoffinvest.Instrument, error)
	GetOrderBook(ctx context.Context, req tinkoffinvest.OrderBookRequest) (*tinkoffinvest.OrderBookResponse, error)

	GetOrderState(ctx context.Context, _ tinkoffinvest.AccountID, _ tinkoffinvest.OrderID) (decimal.Decimal, error)
	CancelOrder(ctx context.Context, accountID tinkoffinvest.AccountID, orderID tinkoffinvest.OrderID) error

	PlaceLimitSellOrder(ctx context.Context, request tinkoffinvest.PlaceOrderRequest) (tinkoffinvest.OrderID, error)
	PlaceLimitBuyOrder(ctx context.Context, request tinkoffinvest.PlaceOrderRequest) (tinkoffinvest.OrderID, error)
}

type ToolsCache interface {
	Get(ctx context.Context, figi tinkoffinvest.FIGI) (toolscache.Tool, error)
}

// Strategy consists in placing two counter orders at the spread border
// with their further adjustment.
type Strategy struct {
	account             tinkoffinvest.AccountID
	ignoreInconsistent  bool
	figis               []tinkoffinvest.FIGI
	minSpreadPercentage float64

	orderPlacer OrderPlacer
	toolsCache  ToolsCache
	logger      zerolog.Logger

	orders      map[tinkoffinvest.FIGI]*ordersPair
	toolConfigs map[tinkoffinvest.FIGI]toolConfig
}

type ordersPair struct {
	toBuy  order
	toSell order
}

type toolConfig struct {
	stocksPerLot int
	minPriceInc  decimal.Decimal
}

type order struct {
	id    tinkoffinvest.OrderID
	price decimal.Decimal
}

func New(
	account tinkoffinvest.AccountID,
	ignoreInconsistent bool,
	minSpreadPercentage float64,
	figis []tinkoffinvest.FIGI,
	orderPlacer OrderPlacer,
	toolsCache ToolsCache,
) (*Strategy, error) {
	s := &Strategy{
		account:             account,
		ignoreInconsistent:  ignoreInconsistent,
		figis:               figis,
		minSpreadPercentage: minSpreadPercentage,
		orderPlacer:         orderPlacer,
		toolsCache:          toolsCache,
		orders:              make(map[tinkoffinvest.FIGI]*ordersPair),
		toolConfigs:         make(map[tinkoffinvest.FIGI]toolConfig),
	}
	s.logger = log.With().Str("strategy", s.Name()).Logger()

	return s, nil
}

func (s *Strategy) Name() string {
	return "spread-parasite"
}

func (s *Strategy) Run(ctx context.Context) error {
	if len(s.figis) == 0 {
		figis, err := s.grepFigisWithEnoughSpread(ctx)
		if err != nil {
			return fmt.Errorf("grep figis: %v", err)
		}

		s.logger.Debug().Msgf("found figis: %#v", figis)
		s.figis = figis
	}

	if err := s.fetchToolConfigs(ctx, s.figis); err != nil {
		return fmt.Errorf("fetch tool configs: %v", err)
	}

	reqs := make([]tinkoffinvest.OrderBookRequest, len(s.figis))
	for i, f := range s.figis {
		reqs[i] = tinkoffinvest.OrderBookRequest{FIGI: f, Depth: orderBookDepth}
		s.orders[f] = new(ordersPair)
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

func (s *Strategy) grepFigisWithEnoughSpread(ctx context.Context) ([]tinkoffinvest.FIGI, error) {
	tools, err := s.orderPlacer.GetTradeAvailableShares(ctx)
	if err != nil {
		return nil, fmt.Errorf("get available instruments: %v", err)
	}

	figis := make([]tinkoffinvest.FIGI, 0, len(tools))
	for _, t := range tools {
		logger := s.logger.With().Str("figi", t.FIGI.S()).Logger()

		orderBook, err := s.orderPlacer.GetOrderBook(ctx, tinkoffinvest.OrderBookRequest{
			FIGI:  t.FIGI,
			Depth: orderBookDepth,
		})
		if err != nil {
			logger.Err(err).Msg("get order book")
			continue
		}

		spread := tinkoffinvest.Spread(orderBook.OrderBook)

		logger.Debug().
			Str("name", t.Name).
			Float64("spread", spread).
			Msg("spread")

		if spread >= s.minSpreadPercentage {
			figis = append(figis, t.FIGI)
		}

		select {
		case <-ctx.Done():
			return nil, nil
		case <-time.After(400 * time.Millisecond):
		}
	}

	return figis[:maxTools], nil
}

func (s *Strategy) fetchToolConfigs(ctx context.Context, figis []tinkoffinvest.FIGI) error {
	s.logger.Debug().Msg("fetch tool info")

	for _, f := range figis {
		tool, err := s.toolsCache.Get(ctx, f)
		if err != nil {
			return fmt.Errorf("get cached tool %v: %v", f, err)
		}
		if tool.MinPriceInc.IsZero() {
			return fmt.Errorf("tool %v: zero min price increment", tool.FIGI)
		}
		if tool.StocksPerLot <= 0 {
			return fmt.Errorf("tool %v: invalid stocks per lot amount", tool.FIGI)
		}

		s.toolConfigs[f] = toolConfig{
			stocksPerLot: tool.StocksPerLot,
			minPriceInc:  tool.MinPriceInc,
		}
	}
	return nil
}

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

	pair := s.orders[change.FIGI]

	if err := s.correctSellOrder(ctx, pair, change, conf, logger); err != nil {
		return fmt.Errorf("correct sell order: %s: %v", change.FIGI, err)
	}

	if err := s.correctBuyOrder(ctx, pair, change, conf, logger); err != nil {
		return fmt.Errorf("correct buy order: %s: %v", change.FIGI, err)
	}

	return nil
}

func (s *Strategy) correctSellOrder(
	ctx context.Context,
	pair *ordersPair,
	change tinkoffinvest.OrderBookChange,
	conf toolConfig,
	logger zerolog.Logger,
) error {
	bestPrice := tinkoffinvest.BestPriceForBuy(change.OrderBook)
	bestPriceGauge.With(l{"best_price_type": bestPriceTypeToBuy, "figi": change.FIGI.S()}).Set(bestPrice.InexactFloat64())

	if orderID := pair.toSell.id; orderID != "" {
		_, err := s.orderPlacer.GetOrderState(ctx, s.account, orderID)
		if !errors.Is(err, tinkoffinvest.ErrOrderWaitExecution) {
			pair.toSell = order{}
		}
	}

	currentPrice := pair.toSell.price
	if ok := currentPrice.IsZero() || currentPrice.GreaterThan(bestPrice); !ok {
		return nil
	}

	bestPrice = bestPrice.Sub(conf.minPriceInc)

	if orderID := pair.toSell.id; orderID != "" {
		if err := s.orderPlacer.CancelOrder(ctx, s.account, orderID); err != nil {
			s.logger.Warn().Str("order_id", orderID.S()).Err(err).Msg("cancel order")
		}
	}

	orderID, err := s.orderPlacer.PlaceLimitSellOrder(ctx, tinkoffinvest.PlaceOrderRequest{
		AccountID: s.account,
		FIGI:      change.FIGI,
		Lots:      lotsInTrade,
		Price:     bestPrice,
	})
	if err != nil {
		if errors.Is(err, tinkoffinvest.ErrNotEnoughStocks) {
			return nil
		}
		return fmt.Errorf("place limit sell order: %v", err)
	}

	common.CollectOrderPrice(bestPrice.InexactFloat64(), s.Name(), change.FIGI, common.OrderTypeLimitSell)
	logger.Info().
		Str("price", bestPrice.String()).
		Str("order_id", orderID.S()).
		Msg("place limit sell order")

	pair.toSell = order{id: orderID, price: bestPrice}
	return nil
}

func (s *Strategy) correctBuyOrder(
	ctx context.Context,
	pair *ordersPair,
	change tinkoffinvest.OrderBookChange,
	conf toolConfig,
	logger zerolog.Logger,
) error {
	bestPrice := tinkoffinvest.BestPriceForSell(change.OrderBook)
	bestPriceGauge.With(l{"best_price_type": bestPriceTypeToSell, "figi": change.FIGI.S()}).Set(bestPrice.InexactFloat64())

	if orderID := pair.toBuy.id; orderID != "" {
		_, err := s.orderPlacer.GetOrderState(ctx, s.account, orderID)
		if !errors.Is(err, tinkoffinvest.ErrOrderWaitExecution) {
			pair.toBuy = order{}
		}
	}

	currentPrice := pair.toBuy.price
	if ok := currentPrice.IsZero() || currentPrice.LessThan(bestPrice); !ok {
		return nil
	}

	bestPrice = bestPrice.Add(conf.minPriceInc)

	if orderID := pair.toBuy.id; orderID != "" {
		if err := s.orderPlacer.CancelOrder(ctx, s.account, orderID); err != nil {
			s.logger.Warn().Str("order_id", orderID.S()).Err(err).Msg("cancel order")
		}
	}

	orderID, err := s.orderPlacer.PlaceLimitBuyOrder(ctx, tinkoffinvest.PlaceOrderRequest{
		AccountID: s.account,
		FIGI:      change.FIGI,
		Lots:      lotsInTrade,
		Price:     bestPrice,
	})
	if err != nil {
		if errors.Is(err, tinkoffinvest.ErrNotEnoughStocks) {
			return nil
		}
		return fmt.Errorf("place limit buy order: %v", err)
	}

	common.CollectOrderPrice(bestPrice.InexactFloat64(), s.Name(), change.FIGI, common.OrderTypeLimitBuy)
	logger.Info().
		Str("price", bestPrice.String()).
		Str("order_id", orderID.S()).
		Msg("place limit buy order")

	pair.toBuy = order{id: orderID, price: bestPrice}
	return nil
}
