package spreadparasite

import (
	"context"

	"github.com/shopspring/decimal"

	"github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest"
)

type OrderPlacer interface {
	GetOrderState(ctx context.Context, _ tinkoffinvest.AccountID, _ tinkoffinvest.OrderID) (decimal.Decimal, error)
	CancelOrder(ctx context.Context, orderID tinkoffinvest.OrderID) error

	PlaceLimitSellOrder(ctx context.Context, request tinkoffinvest.PlaceOrderRequest) (tinkoffinvest.OrderID, error)
	PlaceLimitBuyOrder(ctx context.Context, request tinkoffinvest.PlaceOrderRequest) (tinkoffinvest.OrderID, error)
}

// Strategy consists in placing two counter orders at the spread border
// with their further adjustment.
type Strategy struct {
	account tinkoffinvest.AccountID
	orders  map[tinkoffinvest.FIGI]ordersPair

	orderPlacer OrderPlacer
}

type ordersPair struct {
	toBuy  order
	toSell order
}

type order struct {
	id    tinkoffinvest.OrderID
	price decimal.Decimal
}

func New() (*Strategy, error) {
	return new(Strategy), nil
}

func (s *Strategy) Name() string {
	return "spread-parasite"
}

func (s *Strategy) Run(ctx context.Context) error {
	/*
		tools, err := client.GetTradeAvailableShares(ctx)
		if err != nil {
			return fmt.Errorf("get available instruments: %v", err)
		}

		reqs := make([]tinkoffinvest.OrderBookRequest, 0, len(tools))
		for _, t := range tools {
			logger := log.With().Str("figi", t.FIGI).Logger()

			orderBook, err := client.GetOrderBook(ctx, tinkoffinvest.OrderBookRequest{
				FIGI:  t.FIGI,
				Depth: cfg.Depth,
			})
			if err != nil {
				logger.Err(err).Msg("get order book")
				continue
			}

			spread := tinkoffinvest.Spread(*orderBook)

			logger.Debug().
				Str("name", t.Name).
				Float64("spread", spread).
				Msg("spread")

			if spread >= cfg.MinSpreadPercentage {
				reqs = append(reqs, tinkoffinvest.OrderBookRequest{
					FIGI:  t.FIGI,
					Depth: cfg.Depth,
				})
			}

			select {
			case <-ctx.Done():
				return nil
			case <-time.After(800 * time.Millisecond):
			}
		}

		rand.Shuffle(len(reqs), func(i, j int) {
			reqs[i], reqs[j] = reqs[j], reqs[i]
		})
		reqs = reqs[:cfg.MaxTools]

		strategy, err := spreadmon.New()
		if err != nil {
			return fmt.Errorf("build strategy: %v", err)
		}

		changes, err := client.SubscribeForOrderBookChanges(ctx, reqs)
		if err != nil {
			return fmt.Errorf("subscribe for order book changes: %v", err)
		}

		go func() {
			logger := log.With().Str("strategy", strategy.Name()).Logger()

			select {
			case <-time.After(5 * time.Second):
				logger.Debug().Msg("no order book changes due to period")

			case change, ok := <-changes:
				if !ok {
					return
				}

				func() {
					ctx, cancel := context.WithTimeout(ctx, spreadMonStrategyApplyingTimeout)
					defer cancel()

					if err := s.Apply(ctx, change); err != nil {
						logger.Err(err).Msg("cannot apply order book change")
					}
				}()
			}
		}()
	*/
	return nil
}

func (s *Strategy) Apply(ctx context.Context, change tinkoffinvest.OrderBookChange) error {
	pair := s.orders[change.FIGI]

	{
		bestPrice := tinkoffinvest.BestPriceForBuy(change.OrderBook)
		p := bestPrice.Sub(decimal.RequireFromString("0.01")) // TODO(a.telyshev): Remove hardcore

		if orderID := pair.toSell.id; orderID != "" {
			s.orderPlacer.CancelOrder(ctx, orderID)
		}

		s.orderPlacer.PlaceLimitSellOrder(ctx, tinkoffinvest.PlaceOrderRequest{
			AccountID: s.account,
			FIGI:      change.FIGI,
			Lots:      1,
			Price:     p,
		})
	}

	{
		bestPrice := tinkoffinvest.BestPriceForBuy(change.OrderBook)
		p := bestPrice.Sub(decimal.RequireFromString("0.01")) // TODO(a.telyshev): Remove hardcore

		s.orderPlacer.PlaceLimitSellOrder(ctx, tinkoffinvest.PlaceOrderRequest{
			AccountID: s.account,
			FIGI:      change.FIGI,
			Lots:      1,
			Price:     p,
		})
	}

	return nil
}
