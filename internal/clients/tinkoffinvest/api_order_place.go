package tinkoffinvest

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

var ErrNotEnoughStocks = errors.New("not enough stocks (shares or money)")

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
		return "", fmt.Errorf("grpc post order call: %w", err)
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
		return "", fmt.Errorf("grpc post order call: %w", err)
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
		return "", fmt.Errorf("grpc post order call: %w", err)
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
		return "", fmt.Errorf("grpc post order call: %w", err)
	}

	return OrderID(resp.OrderId), nil
}

func (c *Client) postPbOrder(ctx context.Context, req *investpb.PostOrderRequest) (*investpb.PostOrderResponse, error) {
	ctx = c.auth(ctx)

	var (
		resp *investpb.PostOrderResponse
		err  error
	)
	if c.useSandbox {
		resp, err = c.sandbox.PostSandboxOrder(ctx, req)
	} else {
		resp, err = c.orders.PostOrder(ctx, req)
	}
	if err != nil {
		if isStatusError(err, codes.InvalidArgument, codeNotEnoughAssetsForMarginTrade) {
			log.Warn().Msg("no money no honey")
			return nil, fmt.Errorf("%v: %w", err, ErrNotEnoughStocks)
		}
		return nil, err
	}
	return resp, nil
}
