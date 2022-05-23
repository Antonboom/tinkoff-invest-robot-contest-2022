package tinkoffinvest

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

type currency string

const currencyRUB currency = "rub"

func (c *Client) GetBalance(ctx context.Context, accountID AccountID) (decimal.Decimal, error) {
	ctx = c.auth(ctx)

	if c.useSandbox {
		resp, err := c.sandbox.SandboxPayIn(ctx, &investpb.SandboxPayInRequest{
			AccountId: accountID.S(),
			Amount:    new(investpb.MoneyValue),
		})
		if err != nil {
			return decimal.Zero, fmt.Errorf("grcp sandbox zero payin call: %v", err)
		}
		return newDecimal(resp.Balance.Units, resp.Balance.Nano), nil
	}

	resp, err := c.operations.GetWithdrawLimits(ctx, &investpb.WithdrawLimitsRequest{AccountId: accountID.S()})
	if err != nil {
		return decimal.Zero, fmt.Errorf("grcp get withdraw limits call: %v", err)
	}
	for _, m := range resp.Money {
		if currency(m.Currency) == currencyRUB {
			return newDecimal(m.Units, m.Nano), nil
		}
	}
	return decimal.Zero, nil
}
