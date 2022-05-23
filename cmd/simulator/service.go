package main

import (
	"context"
	stdlog "log"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type Simulator struct {
	investpb.UnimplementedInstrumentsServiceServer
	investpb.UnimplementedMarketDataStreamServiceServer
}

func NewSimulator() *Simulator {
	return new(Simulator)
}

func (s *Simulator) ShareBy(_ context.Context, req *investpb.InstrumentRequest) (*investpb.ShareResponse, error) {
	if req.IdType != investpb.InstrumentIdType_INSTRUMENT_ID_TYPE_FIGI {
		return nil, status.Error(codes.Unimplemented, "simulator supports figis only")
	}
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
			MinPriceIncrement:     &investpb.Quotation{Units: 0, Nano: int32(1 + rand.Intn(10)/100*_10e9)},
			ApiTradeAvailableFlag: true,
		},
	}, nil
}

func (s *Simulator) MarketDataStream(srv investpb.MarketDataStreamService_MarketDataStreamServer) error {
	initialReq, err := srv.Recv()
	if err != nil {
		return err
	}

	obReq, ok := initialReq.Payload.(*investpb.MarketDataRequest_SubscribeOrderBookRequest)
	if !ok {
		return status.Error(codes.Unimplemented, "simulator supports order book only")
	}

	req := obReq.SubscribeOrderBookRequest
	if req.SubscriptionAction != investpb.SubscriptionAction_SUBSCRIPTION_ACTION_SUBSCRIBE {
		return status.Error(codes.Unimplemented, "simulator supports simple subscription only")
	}

	if len(req.Instruments) == 0 {
		return status.Error(codes.InvalidArgument, "no instruments in request")
	}

	subs := make([]*investpb.OrderBookSubscription, len(req.Instruments))
	for i, tool := range req.Instruments {
		subStatus := investpb.SubscriptionStatus_SUBSCRIPTION_STATUS_SUCCESS

		switch tool.Depth {
		case 1, 10, 20, 30, 40, 50:
		default:
			subStatus = investpb.SubscriptionStatus_SUBSCRIPTION_STATUS_DEPTH_IS_INVALID
		}

		subs[i] = &investpb.OrderBookSubscription{
			Figi:               tool.Figi,
			Depth:              tool.Depth,
			SubscriptionStatus: subStatus,
		}
	}

	initialResp := &investpb.MarketDataResponse{
		Payload: &investpb.MarketDataResponse_SubscribeOrderBookResponse{
			SubscribeOrderBookResponse: &investpb.SubscribeOrderBookResponse{
				TrackingId:             uuid.NewString(),
				OrderBookSubscriptions: subs,
			},
		},
	}
	if err := srv.Send(initialResp); err != nil {
		return err
	}

	ctx := srv.Context()

	var wg sync.WaitGroup
	wg.Add(len(req.Instruments))

	for _, tool := range req.Instruments {
		tool := tool

		go func() {
			defer wg.Done()

			for {
				sleep := time.Duration(500+rand.Intn(2000)) * time.Millisecond

				select {
				case <-ctx.Done():
					return

				case <-time.After(sleep):
					resp := &investpb.MarketDataResponse{
						Payload: newRandomOrderBook(tool.Figi, tool.Depth),
					}
					if err := srv.Send(resp); err != nil {
						stdlog.Printf("send order book for %v: %v", tool.Figi, err)
					} else {
						stdlog.Printf("send order book for %v", tool.Figi)
					}
				}
			}
		}()
	}

	wg.Wait()
	return nil
}

const (
	_10e9  = 1_000_000_000
	spread = 0.001 // 0.1 $
)

func newRandomOrderBook(figi string, depth int32) *investpb.MarketDataResponse_Orderbook {
	baseBidUnits := 100 + rand.Int63n(100)
	baseAskUnits := int64(float64(baseBidUnits) * (1. + spread))

	newRandNano := func() int32 { return 1 + rand.Int31n(100)/100*_10e9 }
	newRandQuantity := func() int64 { return 1 + rand.Int63n(100) }

	baseBid := &investpb.Quotation{
		Units: baseBidUnits,
		Nano:  newRandNano(),
	}
	baseAsk := &investpb.Quotation{
		Units: baseAskUnits,
		Nano:  newRandNano(),
	}

	ob := &investpb.MarketDataResponse_Orderbook{
		Orderbook: &investpb.OrderBook{
			Figi:         figi,
			Depth:        depth,
			IsConsistent: true,
			Bids: []*investpb.Order{{
				Price:    baseBid,
				Quantity: newRandQuantity(),
			}},
			Asks: []*investpb.Order{{
				Price:    baseAsk,
				Quantity: newRandQuantity(),
			}},
			Time: timestamppb.Now(),
			LimitUp: &investpb.Quotation{
				Units: baseBidUnits - 50,
				Nano:  newRandNano(),
			},
			LimitDown: &investpb.Quotation{
				Units: baseBidUnits + 50,
				Nano:  newRandNano(),
			},
		},
	}

	for i := 0; i < int((depth-2)/2); i++ {
		ob.Orderbook.Bids = append(ob.Orderbook.Bids, &investpb.Order{
			Price: &investpb.Quotation{
				Units: baseBidUnits,
				Nano:  newRandNano(),
			},
			Quantity: newRandQuantity(),
		})

		ob.Orderbook.Asks = append(ob.Orderbook.Asks, &investpb.Order{
			Price: &investpb.Quotation{
				Units: baseAskUnits,
				Nano:  newRandNano(),
			},
			Quantity: newRandQuantity(),
		})
	}

	return ob
}
