package tinkoffinvest

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

type PlaceOrderRequest struct {
	AccountID AccountID
	FIGI      FIGI
	Lots      int
	Price     decimal.Decimal // For limit orders only.
}

func (c *Client) PlaceMarketSellOrder(ctx context.Context, request PlaceOrderRequest) (OrderID, error) {
	req := &investpb.PostOrderRequest{
		Figi:      request.FIGI.S(),
		Quantity:  int64(request.Lots),
		Direction: investpb.OrderDirection_ORDER_DIRECTION_SELL,
		AccountId: request.AccountID.S(),
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
		Figi:      request.FIGI.S(),
		Quantity:  int64(request.Lots),
		Direction: investpb.OrderDirection_ORDER_DIRECTION_BUY,
		AccountId: request.AccountID.S(),
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
	if request.Price.IsZero() {
		panic("price must be defined for limit order")
	}

	req := &investpb.PostOrderRequest{
		Figi:      request.FIGI.S(),
		Quantity:  int64(request.Lots),
		Price:     adaptDecimalToPbQuotation(request.Price),
		Direction: investpb.OrderDirection_ORDER_DIRECTION_SELL,
		AccountId: request.AccountID.S(),
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
	if request.Price.IsZero() {
		panic("price must be defined for limit order")
	}

	req := &investpb.PostOrderRequest{
		Figi:      request.FIGI.S(),
		Quantity:  int64(request.Lots),
		Price:     adaptDecimalToPbQuotation(request.Price),
		Direction: investpb.OrderDirection_ORDER_DIRECTION_BUY,
		AccountId: request.AccountID.S(),
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
