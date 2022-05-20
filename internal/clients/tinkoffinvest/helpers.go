package tinkoffinvest

import "sort"

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
	if len(ob.Bids) == 0 || len(ob.Acks) == 0 {
		return 0.
	}

	sort.Slice(ob.Bids, func(i, j int) bool {
		return ob.Bids[i].Price.Greater(ob.Bids[j].Price) // Asc by price.
	})
	sort.Slice(ob.Acks, func(i, j int) bool {
		return ob.Acks[i].Price.Less(ob.Acks[j].Price) // Desc by price.
	})

	bestPriceForBuy, bestPriceForSell := ob.Acks[0].Price, ob.Bids[0].Price
	return bestPriceForBuy.Sub(bestPriceForSell).Div(bestPriceForBuy)
}
