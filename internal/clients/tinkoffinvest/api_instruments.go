package tinkoffinvest

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

type Instrument struct {
	FIGI              FIGI
	ISIN              string
	Name              string
	Lot               int
	MinPriceIncrement decimal.Decimal
}

func (c *Client) GetTradeAvailableShares(ctx context.Context) ([]Instrument, error) {
	resp, err := c.instruments.Shares(c.auth(ctx), &investpb.InstrumentsRequest{
		InstrumentStatus: investpb.InstrumentStatus_INSTRUMENT_STATUS_BASE,
	})
	if err != nil {
		return nil, fmt.Errorf("grpc call: %v", err)
	}

	result := make([]Instrument, 0, len(resp.Instruments))
	for _, share := range resp.Instruments {
		av := share.ApiTradeAvailableFlag && share.BuyAvailableFlag && share.SellAvailableFlag
		if !av {
			continue
		}
		result = append(result, adaptPbShareToInstrument(share))
	}
	return result, nil
}

func (c *Client) GetShareByFIGI(ctx context.Context, figi FIGI) (*Instrument, error) {
	resp, err := c.instruments.ShareBy(c.auth(ctx), &investpb.InstrumentRequest{
		IdType: investpb.InstrumentIdType_INSTRUMENT_ID_TYPE_FIGI,
		Id:     figi.S(),
	})
	if err != nil {
		return nil, fmt.Errorf("grcp share by call: %v", err)
	}

	i := adaptPbShareToInstrument(resp.Instrument)
	return &i, nil
}
