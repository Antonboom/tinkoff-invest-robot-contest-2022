package tinkoffinvest //nolint:testpackage

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

func Test_adaptPbQuotationToDecimal(t *testing.T) {
	cases := []struct {
		q        *investpb.Quotation
		expected string
	}{
		{q: nil, expected: "0.00"},
		{q: new(investpb.Quotation), expected: "0.00"},
		{q: &investpb.Quotation{Units: 114, Nano: 250000000}, expected: "114.25"},
		{q: &investpb.Quotation{Units: -200, Nano: -200000000}, expected: "-200.20"},
		{q: &investpb.Quotation{Units: 0, Nano: -10000000}, expected: "-0.01"},
		{q: &investpb.Quotation{Units: 123, Nano: 320000000}, expected: "123.32"},
	}

	for _, tt := range cases {
		t.Run("", func(t *testing.T) {
			d := adaptPbQuotationToDecimal(tt.q)
			assert.Equal(t, tt.expected, d.StringFixed(2))
		})
	}
}

func Test_adaptDecimalToPbQuotation(t *testing.T) {
	cases := []struct {
		d        string
		expected *investpb.Quotation
	}{
		{d: "0", expected: new(investpb.Quotation)},
		{d: "0.00", expected: new(investpb.Quotation)},
		{d: "114.25", expected: &investpb.Quotation{Units: 114, Nano: 250000000}},
		{d: "114.250000000", expected: &investpb.Quotation{Units: 114, Nano: 250000000}},
		{d: "-200.2", expected: &investpb.Quotation{Units: -200, Nano: -200000000}},
		{d: "-200.20", expected: &investpb.Quotation{Units: -200, Nano: -200000000}},
		{d: "-200.200000000", expected: &investpb.Quotation{Units: -200, Nano: -200000000}},
		{d: "-0.01", expected: &investpb.Quotation{Units: 0, Nano: -10000000}},
		{d: "-0.010000000", expected: &investpb.Quotation{Units: 0, Nano: -10000000}},
		{d: "123.32", expected: &investpb.Quotation{Units: 123, Nano: 320000000}},
		{d: "123.320000000", expected: &investpb.Quotation{Units: 123, Nano: 320000000}},
	}

	for _, tt := range cases {
		t.Run(tt.d, func(t *testing.T) {
			q := adaptDecimalToPbQuotation(decimal.RequireFromString(tt.d))
			assert.Equal(t, tt.expected, q)
		})
	}
}

func Test_adaptPbQuotationToDecimal_NullZero(t *testing.T) {
	d := adaptPbQuotationToDecimal(new(investpb.Quotation))
	assert.True(t, d.IsZero())

	d = adaptPbQuotationToDecimal(nil)
	assert.True(t, d.IsZero())

	q := adaptDecimalToPbQuotation(decimal.Zero)
	assert.Equal(t, new(investpb.Quotation), q)
}

func Test_adaptPbOrderbook(t *testing.T) {
	ob := &investpb.OrderBook{
		Figi:         "BBG004RVFFC0",
		Depth:        10,
		IsConsistent: true,
		Bids: []*investpb.Order{
			{Price: &investpb.Quotation{Units: 180, Nano: 520000000}, Quantity: 100},
			{Price: &investpb.Quotation{Units: 180, Nano: 530000000}, Quantity: 11},
			{Price: &investpb.Quotation{Units: 180, Nano: 550000000}, Quantity: 1},
			{Price: &investpb.Quotation{Units: 180, Nano: 600000000}, Quantity: 30},
			{Price: &investpb.Quotation{Units: 180, Nano: 610000000}, Quantity: 278},
		},
		Asks: []*investpb.Order{
			{Price: &investpb.Quotation{Units: 180, Nano: 620000000}, Quantity: 2},
			{Price: &investpb.Quotation{Units: 180, Nano: 650000000}, Quantity: 379},
			{Price: &investpb.Quotation{Units: 180, Nano: 680000000}, Quantity: 100},
			{Price: &investpb.Quotation{Units: 180, Nano: 700000000}, Quantity: 5},
			{Price: &investpb.Quotation{Units: 180, Nano: 770000000}, Quantity: 1},
		},
		Time:      timestamppb.New(time.Unix(1, 1).UTC()),
		LimitUp:   &investpb.Quotation{Units: 200, Nano: 220000000},
		LimitDown: &investpb.Quotation{Units: 95, Nano: 320000000},
	}

	orderBook := adaptPbOrderbook(ob)
	assert.Equal(t, OrderBookChange{
		OrderBook: OrderBook{
			FIGI: "BBG004RVFFC0",
			Bids: []Order{
				{Price: decimal.RequireFromString("180.520000000"), Lots: 100},
				{Price: decimal.RequireFromString("180.530000000"), Lots: 11},
				{Price: decimal.RequireFromString("180.550000000"), Lots: 1},
				{Price: decimal.RequireFromString("180.600000000"), Lots: 30},
				{Price: decimal.RequireFromString("180.610000000"), Lots: 278},
			},
			Asks: []Order{
				{Price: decimal.RequireFromString("180.620000000"), Lots: 2},
				{Price: decimal.RequireFromString("180.650000000"), Lots: 379},
				{Price: decimal.RequireFromString("180.680000000"), Lots: 100},
				{Price: decimal.RequireFromString("180.700000000"), Lots: 5},
				{Price: decimal.RequireFromString("180.770000000"), Lots: 1},
			},
			LimitUp:   decimal.RequireFromString("200.220000000"),
			LimitDown: decimal.RequireFromString("95.320000000"),
		},
		IsConsistent: true,
		FormedAt:     time.Unix(1, 1).UTC(),
	}, orderBook)
}
