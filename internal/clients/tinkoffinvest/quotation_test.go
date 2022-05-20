package tinkoffinvest_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest"
)

func TestQuotation_Mul(t *testing.T) {
	q := newQuotation(125, 12)

	cases := []struct {
		v   float64
		exp tinkoffinvest.Quotation
	}{
		{v: 0, exp: newQuotation(0, 0)},
		{v: 0.5, exp: newQuotation(62, 56)},
		{v: 1, exp: newQuotation(125, 12)},
		{v: 1.01, exp: newQuotation(126, 37)},
		{v: 1.1, exp: newQuotation(137, 63)},
		{v: 3, exp: newQuotation(375, 36)},
	}

	for _, tt := range cases {
		t.Run("", func(t *testing.T) {
			result := q.Mul(tt.v)
			assert.Equal(t, tt.exp, result, "%v * %v = %v != %v", q, tt.v, result, tt.exp)
		})
	}
}

func TestQuotation_Compares(t *testing.T) {
	cases := []struct {
		a, b                tinkoffinvest.Quotation
		expGreater, expLess bool
	}{
		{
			a:          newQuotation(100, 20),
			b:          newQuotation(100, 20),
			expGreater: false,
			expLess:    false,
		},
		{
			a:          newQuotation(100, 20),
			b:          newQuotation(100, 19),
			expGreater: true,
			expLess:    false,
		},
		{
			a:          newQuotation(100, 20),
			b:          newQuotation(99, 20),
			expGreater: true,
			expLess:    false,
		},
		{
			a:          newQuotation(100, 19),
			b:          newQuotation(100, 20),
			expGreater: false,
			expLess:    true,
		},
		{
			a:          newQuotation(99, 20),
			b:          newQuotation(100, 20),
			expGreater: false,
			expLess:    true,
		},
	}

	for _, tt := range cases {
		t.Run("", func(t *testing.T) {
			assert.Equal(t, tt.expGreater, tt.a.Greater(tt.b), "%v not greater than %v", tt.a, tt.b)
			assert.Equal(t, tt.expLess, tt.a.Less(tt.b), "%v not less than %v", tt.a, tt.b)
		})
	}
}

func newQuotation(u, n int) tinkoffinvest.Quotation {
	return tinkoffinvest.Quotation{
		Units: u,
		Nano:  n * 1_000_000_000,
	}
}
