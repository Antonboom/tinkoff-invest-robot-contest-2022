package bullsbearsmon_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest"
	toolscache "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/services/tools-cache"
	bullsbearsmon "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/strategies/bulls-and-bears-mon"
	bullsbearsmonmocks "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/strategies/bulls-and-bears-mon/mocks"
)

const (
	accountID = tinkoffinvest.AccountID("account-xxx")

	figi             = tinkoffinvest.FIGI("BBG004730N88")
	depth            = 1
	dominanceRatio   = 5.5
	profitPercentage = 0.01 // 1%
	stocksPerLot     = 10
)

var d = decimal.RequireFromString

func TestStrategy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orderPlacer := bullsbearsmonmocks.NewMockOrderPlacer(ctrl)
	toolsCache := bullsbearsmonmocks.NewMockToolsCache(ctrl)

	tConfigs := []bullsbearsmon.ToolConfig{
		{
			FIGI:             figi,
			Depth:            depth,
			DominanceRatio:   dominanceRatio,
			ProfitPercentage: profitPercentage, // 1 %
		},
	}

	s, err := bullsbearsmon.New(accountID, false, tConfigs, orderPlacer, toolsCache)
	require.NoError(t, err)

	// Run strategy.

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	toolsCache.EXPECT().Get(gomock.Any(), figi).Return(toolscache.Tool{
		FIGI:         figi,
		StocksPerLot: stocksPerLot,
		MinPriceInc:  d("0.01"),
	}, nil)

	changes := make(chan tinkoffinvest.OrderBookChange)
	orderPlacer.EXPECT().SubscribeForOrderBookChanges(gomock.Any(), []tinkoffinvest.OrderBookRequest{{
		FIGI:  figi,
		Depth: depth,
	}}).Return(changes, nil)

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = s.Run(ctx)
	}()

	// Send order book changes.

	t.Run("not enough ratio", func(t *testing.T) {
		changes <- tinkoffinvest.OrderBookChange{
			OrderBook: tinkoffinvest.OrderBook{
				FIGI: figi,
				Bids: []tinkoffinvest.Order{{
					Price: d("120.330000000"),
					Lots:  100,
				}},
				Asks: []tinkoffinvest.Order{{
					Price: d("120.800000000"),
					Lots:  100,
				}},
				LimitUp:   d("150.200000000"),
				LimitDown: d("90.100000000"),
			},
			IsConsistent: true,
			FormedAt:     time.Now(),
		}
	})

	t.Run("buys are more than sells", func(t *testing.T) {
		oid1 := tinkoffinvest.OrderID("order-1")
		orderPlacer.EXPECT().PlaceMarketBuyOrder(gomock.Any(), tinkoffinvest.PlaceOrderRequest{
			AccountID: accountID,
			FIGI:      figi,
			Lots:      1,
		}).Return(oid1, nil)

		price := d("120.810000000").Mul(decimal.NewFromInt(stocksPerLot))
		orderPlacer.EXPECT().WaitForOrderExecution(gomock.Any(), accountID, oid1).Return(price, nil)

		oid2 := tinkoffinvest.OrderID("order-2")
		orderPlacer.EXPECT().PlaceLimitSellOrder(gomock.Any(), tinkoffinvest.PlaceOrderRequest{
			AccountID: accountID,
			FIGI:      figi,
			Lots:      1,
			Price:     d("122.02"), // 122.0181
		}).Return(oid2, nil)

		changes <- tinkoffinvest.OrderBookChange{
			OrderBook: tinkoffinvest.OrderBook{
				FIGI: figi,
				Bids: []tinkoffinvest.Order{{
					Price: d("120.330000000"),
					Lots:  551,
				}},
				Asks: []tinkoffinvest.Order{{
					Price: d("120.800000000"),
					Lots:  100,
				}},
				LimitUp:   d("150.200000000"),
				LimitDown: d("90.100000000"),
			},
			IsConsistent: true,
			FormedAt:     time.Now(),
		}
	})

	t.Run("sells are more than buys", func(t *testing.T) {
		oid3 := tinkoffinvest.OrderID("order-3")
		orderPlacer.EXPECT().PlaceMarketSellOrder(gomock.Any(), tinkoffinvest.PlaceOrderRequest{
			AccountID: accountID,
			FIGI:      figi,
			Lots:      1,
		}).Return(oid3, nil)

		price := d("120.340000000").Mul(decimal.NewFromInt(stocksPerLot))
		orderPlacer.EXPECT().WaitForOrderExecution(gomock.Any(), accountID, oid3).Return(price, nil)

		oid4 := tinkoffinvest.OrderID("order-4")
		orderPlacer.EXPECT().PlaceLimitBuyOrder(gomock.Any(), tinkoffinvest.PlaceOrderRequest{
			AccountID: accountID,
			FIGI:      figi,
			Lots:      1,
			Price:     d("119.14"), // 119.1366
		}).Return(oid4, nil)

		changes <- tinkoffinvest.OrderBookChange{
			OrderBook: tinkoffinvest.OrderBook{
				FIGI: figi,
				Bids: []tinkoffinvest.Order{{
					Price: d("120.330000000"),
					Lots:  1,
				}},
				Asks: []tinkoffinvest.Order{{
					Price: d("120.800000000"),
					Lots:  6,
				}},
				LimitUp:   d("150.200000000"),
				LimitDown: d("90.100000000"),
			},
			IsConsistent: true,
			FormedAt:     time.Now(),
		}
	})

	// Shutdown.

	cancel()
	<-done
}
