package tinkoffinvest

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

type OrderBookRequest struct {
	FIGI  FIGI
	Depth int
}

type OrderBookResponse struct {
	OrderBook
	LastPrice decimal.Decimal
}

type OrderBook struct {
	FIGI FIGI
	// Bids are orders to buy.
	Bids []Order
	// Asks are orders to sell.
	Asks []Order
	// LimitUp limits buy orders.
	LimitUp decimal.Decimal
	// LimitDown limits sell orders.
	LimitDown decimal.Decimal
}

func (c *Client) GetOrderBook(ctx context.Context, req OrderBookRequest) (*OrderBookResponse, error) {
	resp, err := c.marketData.GetOrderBook(c.auth(ctx), &investpb.GetOrderBookRequest{
		Figi:  req.FIGI.S(),
		Depth: int32(req.Depth),
	})
	if err != nil {
		return nil, fmt.Errorf("grpc get order book call: %v", err)
	}

	return &OrderBookResponse{
		OrderBook: OrderBook{
			FIGI:      FIGI(resp.Figi),
			Bids:      adaptPbOrders(resp.Bids),
			Asks:      adaptPbOrders(resp.Asks),
			LimitUp:   adaptPbQuotationToDecimal(resp.LimitUp),
			LimitDown: adaptPbQuotationToDecimal(resp.LimitDown),
		},
		LastPrice: adaptPbQuotationToDecimal(resp.LastPrice),
	}, nil
}

type OrderBookChange struct {
	OrderBook
	IsConsistent bool
	FormedAt     time.Time
}

type Order struct {
	Price decimal.Decimal
	Lots  int
}

func (c *Client) SubscribeForOrderBookChanges(ctx context.Context, reqs []OrderBookRequest) (<-chan OrderBookChange, error) {
	stream, err := c.marketDataStream.MarketDataStream(c.auth(ctx))
	if err != nil {
		return nil, fmt.Errorf("start grpc stream: %v", err)
	}

	// Send initial request.

	instruments := make([]*investpb.OrderBookInstrument, len(reqs))
	for i, req := range reqs {
		instruments[i] = &investpb.OrderBookInstrument{
			Figi:  req.FIGI.S(),
			Depth: int32(req.Depth), // Overflow impossible.
		}
	}

	// TODO(a.telyshev): Send unsubscribe request in defer?
	if err := stream.Send(&investpb.MarketDataRequest{
		Payload: &investpb.MarketDataRequest_SubscribeOrderBookRequest{
			SubscribeOrderBookRequest: &investpb.SubscribeOrderBookRequest{
				SubscriptionAction: investpb.SubscriptionAction_SUBSCRIPTION_ACTION_SUBSCRIBE,
				Instruments:        instruments,
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("send initial request: %v", err)
	}

	// Receive and validate initial response.

	mdResp, err := stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("recv initial response: %v", err)
	}

	orderBookResp, ok := mdResp.Payload.(*investpb.MarketDataResponse_SubscribeOrderBookResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected response type: %T", mdResp.Payload)
	}

	resp := orderBookResp.SubscribeOrderBookResponse
	logger := log.With().Str("tracking_id", resp.TrackingId).Logger()

	subsMap := make(map[string]*investpb.OrderBookSubscription)
	for _, s := range resp.OrderBookSubscriptions {
		subsMap[s.Figi] = s
	}

	for _, instrument := range instruments {
		s, ok := subsMap[instrument.Figi]
		if !ok {
			return nil, fmt.Errorf(
				"tid %v: figi: %v: no response for requested instrument", resp.TrackingId, instrument.Figi)
		}

		if status := s.SubscriptionStatus; status != investpb.SubscriptionStatus_SUBSCRIPTION_STATUS_SUCCESS {
			return nil, fmt.Errorf(
				"tid %v: figi: %v: unexpected subscription status: %v", resp.TrackingId, instrument.Figi, status)
		}
	}

	// Listen changes.

	changes := make(chan OrderBookChange)
	go func() {
		defer close(changes)

		for {
			resp, err := stream.Recv()
			if err != nil {
				if stream.Context().Err() == nil {
					logger.Err(err).Msg("recv order book change error")
				}
				return
			}

			switch v := resp.Payload.(type) {
			case *investpb.MarketDataResponse_Ping:
				logger.Debug().Msg("order book stream ping")

			case *investpb.MarketDataResponse_Orderbook:
				select {
				case <-ctx.Done():
					return
				case changes <- adaptPbOrderbook(v.Orderbook):
				default:
					// Clients may not have time to process the queue.
				}
			}
		}
	}()
	return changes, nil
}
