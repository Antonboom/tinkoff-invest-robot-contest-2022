package spreadparasite_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest"
	toolscache "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/services/tools-cache"
	spreadparasite "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/strategies/spread-parasite"
	spreadparasitemocks "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/strategies/spread-parasite/mocks"
)

const (
	accountID = tinkoffinvest.AccountID("account-xxx")

	minSpreadPercentage = 0.02 // 2 %
	stocksPerLot        = 10
)

var d = decimal.RequireFromString

var figis = []tinkoffinvest.FIGI{
	"BBG004730N88",
	"BBG000SR0YS4",
}

func TestStrategy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	orderPlacer := spreadparasitemocks.NewMockOrderPlacer(ctrl)
	toolsCache := spreadparasitemocks.NewMockToolsCache(ctrl)

	s, err := spreadparasite.New(accountID, false, minSpreadPercentage, figis, orderPlacer, toolsCache)
	require.NoError(t, err)

	// Run strategy.

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	toolsCache.EXPECT().Get(gomock.Any(), figis[0]).Return(toolscache.Tool{
		FIGI:         figis[0],
		StocksPerLot: stocksPerLot,
		MinPriceInc:  d("0.01"),
	}, nil)

	toolsCache.EXPECT().Get(gomock.Any(), figis[1]).Return(toolscache.Tool{
		FIGI:         figis[1],
		StocksPerLot: stocksPerLot,
		MinPriceInc:  d("5"),
	}, nil)

	changes := make(chan tinkoffinvest.OrderBookChange)
	orderPlacer.EXPECT().SubscribeForOrderBookChanges(gomock.Any(), []tinkoffinvest.OrderBookRequest{
		{FIGI: figis[0], Depth: 1},
		{FIGI: figis[1], Depth: 1},
	}).Return(changes, nil)

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = s.Run(ctx)
	}()

	// Send order book changes.

	oid1 := tinkoffinvest.OrderID("oid1")
	oid2 := tinkoffinvest.OrderID("oid2")

	t.Run("initial change", func(t *testing.T) {
		orderPlacer.EXPECT().PlaceLimitSellOrder(gomock.Any(), tinkoffinvest.PlaceOrderRequest{
			AccountID: accountID,
			FIGI:      figis[0],
			Lots:      1,
			Price:     d("120.790000000"),
		}).Return(oid1, nil)

		orderPlacer.EXPECT().PlaceLimitBuyOrder(gomock.Any(), tinkoffinvest.PlaceOrderRequest{
			AccountID: accountID,
			FIGI:      figis[0],
			Lots:      1,
			Price:     d("120.340000000"),
		}).Return(oid2, nil)

		changes <- tinkoffinvest.OrderBookChange{
			OrderBook: tinkoffinvest.OrderBook{
				FIGI: figis[0],
				Bids: []tinkoffinvest.Order{{
					Price: d("120.330000000"),
					Lots:  12,
				}},
				Asks: []tinkoffinvest.Order{{
					Price: d("120.800000000"),
					Lots:  1,
				}},
			},
			IsConsistent: true,
			FormedAt:     time.Now(),
		}
	})

	oid3 := tinkoffinvest.OrderID("oid3")

	t.Run("new best price for buy", func(t *testing.T) {
		orderPlacer.EXPECT().GetOrderState(gomock.Any(), accountID, oid1).Return(decimal.Zero, tinkoffinvest.ErrOrderWaitExecution)
		orderPlacer.EXPECT().GetOrderState(gomock.Any(), accountID, oid2).Return(decimal.Zero, tinkoffinvest.ErrOrderWaitExecution)

		orderPlacer.EXPECT().CancelOrder(gomock.Any(), accountID, oid2).Return(nil)
		orderPlacer.EXPECT().PlaceLimitBuyOrder(gomock.Any(), tinkoffinvest.PlaceOrderRequest{
			AccountID: accountID,
			FIGI:      figis[0],
			Lots:      1,
			Price:     d("120.360000000"),
		}).Return(oid3, nil)

		changes <- tinkoffinvest.OrderBookChange{
			OrderBook: tinkoffinvest.OrderBook{
				FIGI: figis[0],
				Bids: []tinkoffinvest.Order{{
					Price: d("120.350000000"),
					Lots:  11,
				}},
				Asks: []tinkoffinvest.Order{{
					Price: d("120.790000000"),
					Lots:  2,
				}},
			},
			IsConsistent: true,
			FormedAt:     time.Now(),
		}
	})

	oid4 := tinkoffinvest.OrderID("oid4")

	t.Run("new best price for sell", func(t *testing.T) {
		orderPlacer.EXPECT().GetOrderState(gomock.Any(), accountID, oid1).Return(decimal.Zero, tinkoffinvest.ErrOrderWaitExecution)
		orderPlacer.EXPECT().GetOrderState(gomock.Any(), accountID, oid3).Return(decimal.Zero, tinkoffinvest.ErrOrderWaitExecution)

		orderPlacer.EXPECT().CancelOrder(gomock.Any(), accountID, oid1).Return(nil)
		orderPlacer.EXPECT().PlaceLimitSellOrder(gomock.Any(), tinkoffinvest.PlaceOrderRequest{
			AccountID: accountID,
			FIGI:      figis[0],
			Lots:      1,
			Price:     d("120.770000000"),
		}).Return(oid4, nil)

		changes <- tinkoffinvest.OrderBookChange{
			OrderBook: tinkoffinvest.OrderBook{
				FIGI: figis[0],
				Bids: []tinkoffinvest.Order{{
					Price: d("120.350000000"),
					Lots:  9,
				}},
				Asks: []tinkoffinvest.Order{{
					Price: d("120.780000000"),
					Lots:  3,
				}},
			},
			IsConsistent: true,
			FormedAt:     time.Now(),
		}
	})

	t.Run("no new prices and current orders is alive", func(t *testing.T) {
		orderPlacer.EXPECT().GetOrderState(gomock.Any(), accountID, oid3).Return(decimal.Zero, tinkoffinvest.ErrOrderWaitExecution)
		orderPlacer.EXPECT().GetOrderState(gomock.Any(), accountID, oid4).Return(decimal.Zero, tinkoffinvest.ErrOrderWaitExecution)

		changes <- tinkoffinvest.OrderBookChange{
			OrderBook: tinkoffinvest.OrderBook{
				FIGI: figis[0],
				Bids: []tinkoffinvest.Order{{
					Price: d("120.350000000"),
					Lots:  9,
				}},
				Asks: []tinkoffinvest.Order{{
					Price: d("120.780000000"),
					Lots:  3,
				}},
			},
			IsConsistent: true,
			FormedAt:     time.Now(),
		}
	})

	oid5 := tinkoffinvest.OrderID("oid5")
	oid6 := tinkoffinvest.OrderID("oid6")

	t.Run("no new prices and current orders executed", func(t *testing.T) {
		orderPlacer.EXPECT().GetOrderState(gomock.Any(), accountID, oid3).Return(d("120.350000000"), nil)
		orderPlacer.EXPECT().GetOrderState(gomock.Any(), accountID, oid4).Return(d("120.780000000"), nil)

		orderPlacer.EXPECT().PlaceLimitBuyOrder(gomock.Any(), tinkoffinvest.PlaceOrderRequest{
			AccountID: accountID,
			FIGI:      figis[0],
			Lots:      1,
			Price:     d("120.360000000"),
		}).Return(oid5, nil)

		orderPlacer.EXPECT().PlaceLimitSellOrder(gomock.Any(), tinkoffinvest.PlaceOrderRequest{
			AccountID: accountID,
			FIGI:      figis[0],
			Lots:      1,
			Price:     d("120.770000000"),
		}).Return(oid6, nil)

		changes <- tinkoffinvest.OrderBookChange{
			OrderBook: tinkoffinvest.OrderBook{
				FIGI: figis[0],
				Bids: []tinkoffinvest.Order{{
					Price: d("120.350000000"),
					Lots:  20,
				}},
				Asks: []tinkoffinvest.Order{{
					Price: d("120.780000000"),
					Lots:  21,
				}},
			},
			IsConsistent: true,
			FormedAt:     time.Now(),
		}
	})

	oid7 := tinkoffinvest.OrderID("oid7")
	oid8 := tinkoffinvest.OrderID("oid8")

	t.Run("new best price", func(t *testing.T) {
		orderPlacer.EXPECT().GetOrderState(gomock.Any(), accountID, oid5).Return(decimal.Zero, tinkoffinvest.ErrOrderRejected)
		orderPlacer.EXPECT().GetOrderState(gomock.Any(), accountID, oid6).Return(decimal.Zero, tinkoffinvest.ErrOrderRejected)

		orderPlacer.EXPECT().PlaceLimitBuyOrder(gomock.Any(), tinkoffinvest.PlaceOrderRequest{
			AccountID: accountID,
			FIGI:      figis[0],
			Lots:      1,
			Price:     d("120.370000000"),
		}).Return(oid7, nil)

		orderPlacer.EXPECT().PlaceLimitSellOrder(gomock.Any(), tinkoffinvest.PlaceOrderRequest{
			AccountID: accountID,
			FIGI:      figis[0],
			Lots:      1,
			Price:     d("120.760000000"),
		}).Return(oid8, nil)

		changes <- tinkoffinvest.OrderBookChange{
			OrderBook: tinkoffinvest.OrderBook{
				FIGI: figis[0],
				Bids: []tinkoffinvest.Order{{
					Price: d("120.360000000"),
					Lots:  15,
				}},
				Asks: []tinkoffinvest.Order{{
					Price: d("120.770000000"),
					Lots:  15,
				}},
			},
			IsConsistent: true,
			FormedAt:     time.Now(),
		}
	})

	oid9 := tinkoffinvest.OrderID("oid9")
	oid10 := tinkoffinvest.OrderID("oid10")

	t.Run("other figi change", func(t *testing.T) {
		orderPlacer.EXPECT().PlaceLimitSellOrder(gomock.Any(), tinkoffinvest.PlaceOrderRequest{
			AccountID: accountID,
			FIGI:      figis[1],
			Lots:      1,
			Price:     d("95"),
		}).Return(oid9, nil)

		orderPlacer.EXPECT().PlaceLimitBuyOrder(gomock.Any(), tinkoffinvest.PlaceOrderRequest{
			AccountID: accountID,
			FIGI:      figis[1],
			Lots:      1,
			Price:     d("65"),
		}).Return(oid10, nil)

		changes <- tinkoffinvest.OrderBookChange{
			OrderBook: tinkoffinvest.OrderBook{
				FIGI: figis[1],
				Bids: []tinkoffinvest.Order{{
					Price: d("60"),
					Lots:  55,
				}},
				Asks: []tinkoffinvest.Order{{
					Price: d("100"),
					Lots:  66,
				}},
			},
			IsConsistent: true,
			FormedAt:     time.Now(),
		}
	})

	// Shutdown.

	cancel()
	<-done
}
