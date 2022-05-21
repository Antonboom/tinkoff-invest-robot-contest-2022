package tinkoffinvest //nolint:testpackage

import (
	"testing"

	"github.com/stretchr/testify/assert"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

func Test_adaptPbQuotationToDecimal_AndVisaVersa(t *testing.T) {
	cases := []struct {
		units, nano int
		expected    string
	}{
		{units: 114, nano: 250000000, expected: "114.25"},
		{units: -200, nano: -200000000, expected: "-200.20"},
		{units: 0, nano: -10000000, expected: "-0.01"},
		{units: 123, nano: 320000000, expected: "123.32"},
	}

	for _, tt := range cases {
		t.Run("", func(t *testing.T) {
			q := &investpb.Quotation{
				Units: int64(tt.units),
				Nano:  int32(tt.nano),
			}
			d := adaptPbQuotationToDecimal(q)
			assert.Equal(t, tt.expected, d.StringFixed(2))

			q2 := adaptDecimalToPbQuotation(d)
			assert.Equal(t, q, q2)
		})
	}
}
