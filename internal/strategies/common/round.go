package common

import (
	"fmt"

	"github.com/shopspring/decimal"
)

func RoundToMinPriceIncrement(price, step decimal.Decimal) decimal.Decimal {
	if step.IsNegative() {
		panic(fmt.Sprintf("invalid usage of common.RoundToMinPriceIncrement: step: %s < 0", step.String()))
	}
	return price.Div(step).Round(0).Mul(step)
}
