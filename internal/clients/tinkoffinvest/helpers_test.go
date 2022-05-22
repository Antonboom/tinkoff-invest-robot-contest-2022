package tinkoffinvest_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func TestCountLots(t *testing.T) {
	cases := []struct {
		orders  []tinkoffinvest.Order
		expLots int
	}{
		{
			orders:  nil,
			expLots: 0,
		},
		{
			orders:  []tinkoffinvest.Order{},
			expLots: 0,
		},
		{
			orders:  []tinkoffinvest.Order{{Lots: 1}},
			expLots: 1,
		},
		{
			orders:  []tinkoffinvest.Order{{Lots: 1}, {Lots: 2}, {Lots: 3}},
			expLots: 6,
		},
		{
			orders: []tinkoffinvest.Order{
				{Price: decimal.RequireFromString("232.330000000"), Lots: 240},
				{Price: decimal.RequireFromString("232.340000000"), Lots: 988},
				{Price: decimal.RequireFromString("232.350000000"), Lots: 240},
				{Price: decimal.RequireFromString("232.360000000"), Lots: 4},
				{Price: decimal.RequireFromString("232.380000000"), Lots: 41},
			},
			expLots: 1513,
		},
	}

	for _, tt := range cases {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.expLots, tinkoffinvest.CountLots(tt.orders))
		})
	}
}

func TestSpread(t *testing.T) {
	cases := []struct {
		expSpread float64
		book      tinkoffinvest.OrderBook
	}{
		{
			expSpread: 0.000129,
			book: tinkoffinvest.OrderBook{
				FIGI: "BBG004730RP0",
				Bids: []tinkoffinvest.Order{
					{Price: decimal.RequireFromString("232.330000000"), Lots: 240},
					{Price: decimal.RequireFromString("232.340000000"), Lots: 988},
					{Price: decimal.RequireFromString("232.350000000"), Lots: 240},
					{Price: decimal.RequireFromString("232.360000000"), Lots: 4},
					{Price: decimal.RequireFromString("232.380000000"), Lots: 41},
				},
				Acks: []tinkoffinvest.Order{
					{Price: decimal.RequireFromString("232.410000000"), Lots: 166},
					{Price: decimal.RequireFromString("232.420000000"), Lots: 112},
					{Price: decimal.RequireFromString("232.450000000"), Lots: 244},
					{Price: decimal.RequireFromString("232.480000000"), Lots: 314},
					{Price: decimal.RequireFromString("232.490000000"), Lots: 1},
				},
			},
		},
		{
			expSpread: 0.000055,
			book: tinkoffinvest.OrderBook{
				FIGI: "BBG004RVFFC0",
				Bids: []tinkoffinvest.Order{
					{Price: decimal.RequireFromString("180.520000000"), Lots: 100},
					{Price: decimal.RequireFromString("180.530000000"), Lots: 11},
					{Price: decimal.RequireFromString("180.550000000"), Lots: 1},
					{Price: decimal.RequireFromString("180.600000000"), Lots: 30},
					{Price: decimal.RequireFromString("180.610000000"), Lots: 278},
				},
				Acks: []tinkoffinvest.Order{
					{Price: decimal.RequireFromString("180.620000000"), Lots: 2},
					{Price: decimal.RequireFromString("180.650000000"), Lots: 379},
					{Price: decimal.RequireFromString("180.680000000"), Lots: 100},
					{Price: decimal.RequireFromString("180.700000000"), Lots: 5},
					{Price: decimal.RequireFromString("180.770000000"), Lots: 1},
				},
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.book.FIGI.S(), func(t *testing.T) {
			ob := tt.book
			rand.Shuffle(len(ob.Acks), func(i, j int) { ob.Bids[i], ob.Bids[j] = ob.Bids[j], ob.Bids[i] })
			rand.Shuffle(len(ob.Acks), func(i, j int) { ob.Acks[i], ob.Acks[j] = ob.Acks[j], ob.Acks[i] })

			assert.InDelta(t, tt.expSpread, tinkoffinvest.Spread(ob), 0.000001)
		})
	}
}

func TestSpread_NoOrders(t *testing.T) {
	assert.Equal(t, 0.0, tinkoffinvest.Spread(tinkoffinvest.OrderBook{Bids: []tinkoffinvest.Order{{Lots: 1}}}))
	assert.Equal(t, 0.0, tinkoffinvest.Spread(tinkoffinvest.OrderBook{Acks: []tinkoffinvest.Order{{Lots: 1}}}))
	assert.Equal(t, 0.0, tinkoffinvest.Spread(tinkoffinvest.OrderBook{}))
}
