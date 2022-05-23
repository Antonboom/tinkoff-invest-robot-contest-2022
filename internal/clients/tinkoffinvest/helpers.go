package tinkoffinvest

import (
	"sort"

	"github.com/shopspring/decimal"
)

// CountLots counts sum of orders' lots.
func CountLots(orders []Order) int {
	var result int
	for _, o := range orders {
		result += o.Lots
	}
	return result
}

// Spread returns order book spread as percentage. May be negative.
// Can affect OrderBook.Bids & OrderBook.Asks slices order.
func Spread(ob OrderBook) float64 {
	bestPriceForBuy, bestPriceForSell := BestPriceForBuy(ob), BestPriceForSell(ob)
	if bestPriceForBuy.IsZero() || bestPriceForSell.IsZero() {
		return 0.
	}
	return bestPriceForBuy.Sub(bestPriceForSell).Div(bestPriceForBuy).InexactFloat64()
}

// BestPriceForBuy returns the lowest price among sell orders.
func BestPriceForBuy(ob OrderBook) decimal.Decimal {
	if len(ob.Asks) == 0 {
		return decimal.Zero
	}

	sort.Slice(ob.Asks, func(i, j int) bool {
		return ob.Asks[i].Price.LessThan(ob.Asks[j].Price) // Asc by price.
	})
	return ob.Asks[0].Price
}

// BestPriceForSell returns the highest price among buy orders.
func BestPriceForSell(ob OrderBook) decimal.Decimal {
	if len(ob.Bids) == 0 {
		return decimal.Zero
	}

	sort.Slice(ob.Bids, func(i, j int) bool {
		return ob.Bids[i].Price.GreaterThan(ob.Bids[j].Price) // Desc by price.
	})
	return ob.Bids[0].Price
}
