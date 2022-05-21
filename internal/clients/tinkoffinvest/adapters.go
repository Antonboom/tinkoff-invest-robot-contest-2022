package tinkoffinvest

import (
	"github.com/shopspring/decimal"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

const _10e9 = 1_000_000_000.

func adaptPbQuotationToDecimal(q *investpb.Quotation) decimal.Decimal {
	if q == nil {
		return decimal.Zero
	}
	return newDecimal(q.Units, int64(q.Nano))
}

func newDecimal(units, nano int64) decimal.Decimal {
	if units == 0 && nano == 0 {
		return decimal.Zero
	}
	return decimal.New(units*_10e9+nano, -9)
}

var one = decimal.NewFromInt(1)

func adaptDecimalToPbQuotation(d decimal.Decimal) *investpb.Quotation {
	return &investpb.Quotation{
		Units: d.IntPart(),
		Nano:  int32(d.Mod(one).CoefficientInt64()), // Possible overflow.
	}
}

func adaptPbOrderbook(ob *investpb.OrderBook) OrderBookChange {
	return OrderBookChange{
		OrderBook: OrderBook{
			FIGI:      FIGI(ob.Figi),
			Bids:      adaptPbOrders(ob.Bids),
			Acks:      adaptPbOrders(ob.Asks),
			LimitUp:   adaptPbQuotationToDecimal(ob.LimitUp),
			LimitDown: adaptPbQuotationToDecimal(ob.LimitDown),
		},
		IsConsistent: ob.IsConsistent,
		FormedAt:     ob.Time.AsTime(),
	}
}

func adaptPbOrders(orders []*investpb.Order) []Order {
	result := make([]Order, 0, len(orders))
	for _, o := range orders {
		result = append(result, adaptPbOrder(o))
	}
	return result
}

func adaptPbOrder(o *investpb.Order) Order {
	return Order{
		Price: adaptPbQuotationToDecimal(o.Price),
		Lots:  int(o.Quantity), // Possible overflow.
	}
}

func adaptPbShareToInstrument(share *investpb.Share) Instrument {
	return Instrument{
		FIGI:              FIGI(share.Figi),
		ISIN:              share.Isin,
		Name:              share.Name,
		Lot:               int(share.Lot),
		MinPriceIncrement: adaptPbQuotationToDecimal(share.MinPriceIncrement).String(),
	}
}
