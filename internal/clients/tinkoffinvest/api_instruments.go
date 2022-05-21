package tinkoffinvest

import (
	"context"
	"fmt"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

type Instrument struct {
	FIGI              string
	ISIN              string
	Name              string
	Lot               int
	MinPriceIncrement string
}

func (c *Client) GetTradeAvailableShares(ctx context.Context) ([]Instrument, error) {
	resp, err := c.instruments.Shares(c.auth(ctx), &investpb.InstrumentsRequest{
		InstrumentStatus: investpb.InstrumentStatus_INSTRUMENT_STATUS_BASE,
	})
	if err != nil {
		return nil, fmt.Errorf("grpc call: %v", err)
	}

	result := make([]Instrument, 0, len(resp.Instruments))
	for _, bond := range resp.Instruments {
		av := bond.ApiTradeAvailableFlag && bond.BuyAvailableFlag && bond.SellAvailableFlag
		if !av {
			continue
		}

		result = append(result, Instrument{
			FIGI:              bond.Figi,
			ISIN:              bond.Isin,
			Name:              bond.Name,
			Lot:               int(bond.Lot),
			MinPriceIncrement: adaptPbQuotationToDecimal(bond.MinPriceIncrement).String(),
		})
	}
	return result, nil
}
