package tinkoffinvest

func CountLots(orders []Order) int {
	var result int
	for _, o := range orders {
		result += o.Lots
	}
	return result
}
