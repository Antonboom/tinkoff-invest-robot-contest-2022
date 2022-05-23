package tinkoffinvest

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

type instrumentType string

const instrumentTypeShares instrumentType = "share"

type Portfolio struct {
	TotalSharesPrice decimal.Decimal
	Shares           []PortfolioPosition
}

type PortfolioPosition struct {
	FIGI     FIGI
	Quantity int
	AvgPrice decimal.Decimal
}

func (c *Client) GetPortfolio(ctx context.Context, accountID AccountID) (*Portfolio, error) {
	ctx = c.auth(ctx)

	req := &investpb.PortfolioRequest{
		AccountId: accountID.S(),
	}

	var (
		resp *investpb.PortfolioResponse
		err  error
	)
	if c.useSandbox {
		resp, err = c.sandbox.GetSandboxPortfolio(ctx, req)
	} else {
		resp, err = c.operations.GetPortfolio(ctx, req)
	}
	if err != nil {
		return nil, fmt.Errorf("grpc get portfolio call: %v", err)
	}

	shares := make([]PortfolioPosition, 0, len(resp.Positions))
	for _, p := range resp.Positions {
		if instrumentType(p.InstrumentType) == instrumentTypeShares {
			shares = append(shares, PortfolioPosition{
				FIGI:     FIGI(p.Figi),
				Quantity: int(adaptPbQuotationToDecimal(p.Quantity).IntPart()),
				AvgPrice: newDecimal(p.AveragePositionPrice.Units, p.AveragePositionPrice.Nano),
			})
		}
	}

	return &Portfolio{
		TotalSharesPrice: newDecimal(resp.TotalAmountShares.Units, resp.TotalAmountShares.Nano),
		Shares:           shares,
	}, nil
}
