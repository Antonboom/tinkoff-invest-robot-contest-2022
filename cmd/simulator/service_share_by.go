package main

import (
	"context"
	"math/rand"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

func (s *Simulator) ShareBy(_ context.Context, req *investpb.InstrumentRequest) (*investpb.ShareResponse, error) {
	if req.IdType != investpb.InstrumentIdType_INSTRUMENT_ID_TYPE_FIGI {
		return nil, status.Error(codes.Unimplemented, "simulator supports figis only")
	}

	minPriceIncNano := int32(float32(1.+rand.Int31n(10)) / 100. * _10e9)
	return &investpb.ShareResponse{
		Instrument: &investpb.Share{
			Figi:                  req.Id,
			Isin:                  "simulator",
			Lot:                   1 + rand.Int31n(10),
			Currency:              "rub",
			Name:                  "simulator",
			TradingStatus:         investpb.SecurityTradingStatus_SECURITY_TRADING_STATUS_NORMAL_TRADING,
			BuyAvailableFlag:      true,
			SellAvailableFlag:     true,
			MinPriceIncrement:     &investpb.Quotation{Units: 0, Nano: minPriceIncNano},
			ApiTradeAvailableFlag: true,
		},
	}, nil
}
