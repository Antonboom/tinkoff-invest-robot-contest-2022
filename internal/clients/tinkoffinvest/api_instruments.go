package tinkoffinvest

import (
	"context"
	"fmt"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

type Instrument struct {
	FIGI string
	ISIN string
	Name string
}

func (c *Client) GetInstruments(ctx context.Context) ([]Instrument, error) {
	resp, err := c.instruments.Bonds(c.auth(ctx), &investpb.InstrumentsRequest{
		InstrumentStatus: investpb.InstrumentStatus_INSTRUMENT_STATUS_BASE,
	})
	if err != nil {
		return nil, fmt.Errorf("grpc call: %v", err)
	}

	result := make([]Instrument, len(resp.Instruments))
	for i, instr := range resp.Instruments {
		result[i] = Instrument{
			FIGI: instr.Figi,
			ISIN: instr.Isin,
			Name: instr.Name,
		}
	}
	return result, nil
}
