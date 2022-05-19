package tinkoffinvest_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest"
)

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
				{Lots: 1},
				{Lots: 2},
				{Lots: 3},
				{Lots: 1},
				{Lots: 2},
				{Lots: 3},
				{Lots: 1},
				{Lots: 2},
				{Lots: 3},
				{Lots: 1},
				{Lots: 2},
				{Lots: 3},
				{Lots: 0},
			},
			expLots: 6 * 4,
		},
	}

	for _, tt := range cases {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.expLots, tinkoffinvest.CountLots(tt.orders))
		})
	}
}
