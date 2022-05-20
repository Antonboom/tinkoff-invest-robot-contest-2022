package tinkoffinvest

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

type OrderBookRequest struct {
	FIGI  string
	Depth int
}

type OrderBook struct {
	FIGI      string
	Bids      []Order
	Acks      []Order
	LastPrice Quotation
	// LimitUp limits buy orders.
	LimitUp Quotation
	// LimitDown limits sell orders.
	LimitDown Quotation
}

func (c *Client) GetOrderBook(ctx context.Context, req OrderBookRequest) (*OrderBook, error) {
	resp, err := c.marketData.GetOrderBook(c.auth(ctx), &investpb.GetOrderBookRequest{
		Figi:  req.FIGI,
		Depth: int32(req.Depth),
	})
	if err != nil {
		return nil, fmt.Errorf("grpc get order book call: %v", err)
	}

	return &OrderBook{
		FIGI:      resp.Figi,
		Bids:      adaptPbOrders(resp.Bids),
		Acks:      adaptPbOrders(resp.Asks),
		LastPrice: adaptPbQuotation(resp.LastPrice),
		LimitUp:   adaptPbQuotation(resp.LimitUp),
		LimitDown: adaptPbQuotation(resp.LimitDown),
	}, nil
}

type OrderBookChange struct {
	FIGI         string
	IsConsistent bool
	Bids         []Order
	Acks         []Order
	FormedAt     time.Time
	// LimitUp limits buy orders.
	LimitUp Quotation
	// LimitDown limits sell orders.
	LimitDown Quotation
}

type Order struct {
	Price Quotation
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
			Figi:  req.FIGI,
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
				"tid %v: figi: %v: no response for requested instrumen", resp.TrackingId, instrument.Figi)
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

func adaptPbOrderbook(ob *investpb.OrderBook) OrderBookChange {
	return OrderBookChange{
		FIGI:         ob.Figi,
		IsConsistent: ob.IsConsistent,
		Bids:         adaptPbOrders(ob.Bids),
		Acks:         adaptPbOrders(ob.Asks),
		LimitUp:      adaptPbQuotation(ob.LimitUp),
		LimitDown:    adaptPbQuotation(ob.LimitDown),
		FormedAt:     ob.Time.AsTime(),
	}
}

func adaptPbOrders(orders []*investpb.Order) []Order {
	result := make([]Order, 0, len(orders))
	for _, o := range orders {
		result = append(result, adaptPbOrder(o))
	}
	return result
}

func adaptPbOrder(o *investpb.Order) Order {
	return Order{
		Price: adaptPbQuotation(o.Price),
		Lots:  int(o.Quantity), // Overflow impossible.
	}
}

func adaptPbQuotation(q *investpb.Quotation) Quotation {
	if q == nil {
		return Quotation{}
	}

	// Overflows impossible.
	return Quotation{
		Units: int(q.Units),
		Nano:  int(q.Nano),
	}
}
