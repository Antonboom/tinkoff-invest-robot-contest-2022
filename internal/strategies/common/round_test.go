package common_test

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/strategies/common"
)

func TestRoundToMinPriceIncrement(t *testing.T) {
	cases := []struct {
		in, step, expected string
	}{
		{in: "122.018100000", step: "0.01", expected: "122.02"},
		{in: "3.462300000", step: "0.0001", expected: "3.4623"},
		{in: "3.460000000", step: "0.0001", expected: "3.46"},
		{in: "5.038000000", step: "0.2", expected: "5"},
		{in: "5.438000000", step: "0.2", expected: "5.4"},
		{in: "5.038000000", step: "0.02", expected: "5.04"},
		{in: "130.272700000", step: "0.5", expected: "130.5"},
		{in: "130.272700001", step: "0.1", expected: "130.3"},
		{in: "141.891400000", step: "0.0005", expected: "141.8915"},
		{in: "201.01", step: "5", expected: "200"},
		{in: "204", step: "5", expected: "205"},
		{in: "202.02", step: "10", expected: "200"},
		{in: "207.02", step: "10", expected: "210"},
		{in: "203.03", step: "50", expected: "200"},
		{in: "225.01", step: "50", expected: "250"},
	}

	for _, tt := range cases {
		t.Run(tt.in, func(t *testing.T) {
			in, step := decimal.RequireFromString(tt.in), decimal.RequireFromString(tt.step)
			rounded := common.RoundToMinPriceIncrement(in, step)
			assert.Equal(t, tt.expected, rounded.String())
			assert.True(t, rounded.Mod(step).IsZero())
		})
	}
}
