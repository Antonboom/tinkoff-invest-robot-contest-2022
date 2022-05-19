package tinkoffinvest

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

type PlaceOrderRequest struct {
	AccountID AccountID
	FIGI      string
	Lots      int
	Price     *Quotation // For limit orders only.
}

func (c *Client) PlaceMarketSellOrder(ctx context.Context, request PlaceOrderRequest) (OrderID, error) {
	req := &investpb.PostOrderRequest{
		Figi:      request.FIGI,
		Quantity:  int64(request.Lots),
		Direction: investpb.OrderDirection_ORDER_DIRECTION_SELL,
		AccountId: string(request.AccountID),
		OrderType: investpb.OrderType_ORDER_TYPE_MARKET,
		OrderId:   uuid.New().String(),
	}

	resp, err := c.postPbOrder(ctx, req)
	if err != nil {
		return "", fmt.Errorf("grpc post order call: %v", err)
	}

	return OrderID(resp.OrderId), nil
}

func (c *Client) PlaceMarketBuyOrder(ctx context.Context, request PlaceOrderRequest) (OrderID, error) {
	req := &investpb.PostOrderRequest{
		Figi:      request.FIGI,
		Quantity:  int64(request.Lots),
		Direction: investpb.OrderDirection_ORDER_DIRECTION_BUY,
		AccountId: string(request.AccountID),
		OrderType: investpb.OrderType_ORDER_TYPE_MARKET,
		OrderId:   uuid.New().String(),
	}

	resp, err := c.postPbOrder(ctx, req)
	if err != nil {
		return "", fmt.Errorf("grpc post order call: %v", err)
	}

	return OrderID(resp.OrderId), nil
}

func (c *Client) PlaceLimitSellOrder(ctx context.Context, request PlaceOrderRequest) (OrderID, error) {
	if request.Price == nil {
		panic("price must be defined for limit order")
	}

	req := &investpb.PostOrderRequest{
		Figi:      request.FIGI,
		Quantity:  int64(request.Lots),
		Price:     adaptQuotationToPb(*request.Price),
		Direction: investpb.OrderDirection_ORDER_DIRECTION_SELL,
		AccountId: string(request.AccountID),
		OrderType: investpb.OrderType_ORDER_TYPE_LIMIT,
		OrderId:   uuid.New().String(),
	}
	resp, err := c.postPbOrder(ctx, req)
	if err != nil {
		return "", fmt.Errorf("grpc post order call: %v", err)
	}

	return OrderID(resp.OrderId), nil
}

func (c *Client) PlaceLimitBuyOrder(ctx context.Context, request PlaceOrderRequest) (OrderID, error) {
	if request.Price == nil {
		panic("price must be defined for limit order")
	}

	req := &investpb.PostOrderRequest{
		Figi:      request.FIGI,
		Quantity:  int64(request.Lots),
		Price:     adaptQuotationToPb(*request.Price),
		Direction: investpb.OrderDirection_ORDER_DIRECTION_BUY,
		AccountId: string(request.AccountID),
		OrderType: investpb.OrderType_ORDER_TYPE_LIMIT,
		OrderId:   uuid.New().String(),
	}

	resp, err := c.postPbOrder(ctx, req)
	if err != nil {
		return "", fmt.Errorf("grpc post order call: %v", err)
	}

	return OrderID(resp.OrderId), nil
}

func (c *Client) postPbOrder(ctx context.Context, req *investpb.PostOrderRequest) (*investpb.PostOrderResponse, error) {
	ctx = c.auth(ctx)

	if c.useSandbox {
		return c.sandbox.PostSandboxOrder(ctx, req)
	}
	return c.orders.PostOrder(ctx, req)
}

func adaptQuotationToPb(q Quotation) *investpb.Quotation {
	return &investpb.Quotation{
		Units: int64(q.Units),
		Nano:  int32(q.Nano),
	}
}
