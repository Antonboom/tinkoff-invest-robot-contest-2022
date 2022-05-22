package toolscache_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest"
	toolscache "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/services/tools-cache"
	toolscachemocks "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/services/tools-cache/mocks"
)

func TestCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	investClient := toolscachemocks.NewMockSharesProvider(ctrl)
	c := toolscache.New(investClient)

	const f1, f2 = tinkoffinvest.FIGI("f1"), tinkoffinvest.FIGI("f2")

	t.Run("no extra api call", func(t *testing.T) {
		investClient.EXPECT().GetShareByFIGI(gomock.Any(), f1).Return(
			&tinkoffinvest.Instrument{
				FIGI:              f1,
				Lot:               10,
				MinPriceIncrement: decimal.RequireFromString("100.10"),
			}, nil)

		expf1Tool := toolscache.Tool{
			FIGI:         f1,
			StocksPerLot: 10,
			MinPriceInc:  decimal.RequireFromString("100.10"),
		}

		tool, err := c.Get(context.Background(), f1)
		require.NoError(t, err)
		assert.Equal(t, expf1Tool, tool)

		tool, err = c.Get(context.Background(), f1)
		require.NoError(t, err)
		assert.Equal(t, expf1Tool, tool)
	})

	t.Run("fetch unknown figi", func(t *testing.T) {
		investClient.EXPECT().GetShareByFIGI(gomock.Any(), f2).Return(
			&tinkoffinvest.Instrument{
				FIGI:              f2,
				Lot:               20,
				MinPriceIncrement: decimal.RequireFromString("200.20"),
			}, nil)

		expf2Tool := toolscache.Tool{
			FIGI:         f2,
			StocksPerLot: 20,
			MinPriceInc:  decimal.RequireFromString("200.20"),
		}

		tool, err := c.Get(context.Background(), f2)
		require.NoError(t, err)
		assert.Equal(t, expf2Tool, tool)

		tool, err = c.Get(context.Background(), f2)
		require.NoError(t, err)
		assert.Equal(t, expf2Tool, tool)

		tool, err = c.Get(context.Background(), f1)
		require.NoError(t, err)
		assert.Equal(t, toolscache.Tool{
			FIGI:         f1,
			StocksPerLot: 10,
			MinPriceInc:  decimal.RequireFromString("100.10"),
		}, tool)
	})
}
