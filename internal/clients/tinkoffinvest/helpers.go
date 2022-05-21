package tinkoffinvest

import (
	"sort"

	"github.com/shopspring/decimal"
)

func CountLots(orders []Order) int {
	var result int
	for _, o := range orders {
		result += o.Lots
	}
	return result
}

// Spread returns order book spread as percentage. May be negative.
// Can affect bids & acks slices order.
func Spread(ob OrderBook) float64 {
	bestPriceForBuy, bestPriceForSell := BestPriceForBuy(ob), BestPriceForSell(ob)
	if bestPriceForBuy.IsZero() || bestPriceForSell.IsZero() {
		return 0.
	}
	return bestPriceForBuy.Sub(bestPriceForSell).Div(bestPriceForBuy).InexactFloat64()
}

func BestPriceForBuy(ob OrderBook) decimal.Decimal {
	if len(ob.Acks) == 0 {
		return decimal.Zero
	}

	sort.Slice(ob.Acks, func(i, j int) bool {
		return ob.Acks[i].Price.LessThan(ob.Acks[j].Price) // Asc by price.
	})
	return ob.Acks[0].Price
}

func BestPriceForSell(ob OrderBook) decimal.Decimal {
	if len(ob.Bids) == 0 {
		return decimal.Zero
	}

	sort.Slice(ob.Bids, func(i, j int) bool {
		return ob.Bids[i].Price.GreaterThan(ob.Bids[j].Price) // Desc by price.
	})
	return ob.Bids[0].Price
}
