package tinkoffinvest

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

var (
	ErrOrderRejected  = errors.New("order rejected")
	ErrOrderCancelled = errors.New("order cancelled by user")
)

type PlaceOrderRequest struct {
	AccountID string
	FIGI      string
	Lots      int
	Price     *Quotation // For limit orders only.
}

type PlaceOrderResponse struct {
	OrderID       string
	ExecutedPrice Quotation
}

func (c *Client) PlaceMarketSellOrder(ctx context.Context, request PlaceOrderRequest) (*PlaceOrderResponse, error) {
	req := &investpb.PostOrderRequest{
		Figi:      request.FIGI,
		Quantity:  int64(request.Lots),
		Direction: investpb.OrderDirection_ORDER_DIRECTION_SELL,
		AccountId: request.AccountID,
		OrderType: investpb.OrderType_ORDER_TYPE_MARKET,
		OrderId:   uuid.New().String(),
	}

	resp, err := c.postPbOrder(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("grpc post order call: %v", err)
	}

	return adaptPbPostOrderResponse(resp), nil
}

func (c *Client) PlaceMarketBuyOrder(ctx context.Context, request PlaceOrderRequest) (*PlaceOrderResponse, error) {
	req := &investpb.PostOrderRequest{
		Figi:      request.FIGI,
		Quantity:  int64(request.Lots),
		Direction: investpb.OrderDirection_ORDER_DIRECTION_BUY,
		AccountId: request.AccountID,
		OrderType: investpb.OrderType_ORDER_TYPE_MARKET,
		OrderId:   uuid.New().String(),
	}

	resp, err := c.postPbOrder(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("grpc post order call: %v", err)
	}

	return adaptPbPostOrderResponse(resp), nil
}

func (c *Client) PlaceLimitSellOrder(ctx context.Context, request PlaceOrderRequest) (*PlaceOrderResponse, error) {
	if request.Price == nil {
		panic("price must be defined for limit order")
	}

	req := &investpb.PostOrderRequest{
		Figi:      request.FIGI,
		Quantity:  int64(request.Lots),
		Price:     adaptQuotationToPb(*request.Price),
		Direction: investpb.OrderDirection_ORDER_DIRECTION_SELL,
		AccountId: request.AccountID,
		OrderType: investpb.OrderType_ORDER_TYPE_LIMIT,
		OrderId:   uuid.New().String(),
	}
	resp, err := c.postPbOrder(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("grpc post order call: %v", err)
	}

	return adaptPbPostOrderResponse(resp), nil
}

func (c *Client) PlaceLimitBuyOrder(ctx context.Context, request PlaceOrderRequest) (*PlaceOrderResponse, error) {
	if request.Price == nil {
		panic("price must be defined for limit order")
	}

	req := &investpb.PostOrderRequest{
		Figi:      request.FIGI,
		Quantity:  int64(request.Lots),
		Price:     adaptQuotationToPb(*request.Price),
		Direction: investpb.OrderDirection_ORDER_DIRECTION_BUY,
		AccountId: request.AccountID,
		OrderType: investpb.OrderType_ORDER_TYPE_LIMIT,
		OrderId:   uuid.New().String(),
	}

	resp, err := c.postPbOrder(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("grpc post order call: %v", err)
	}

	return adaptPbPostOrderResponse(resp), nil
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
		return nil, err
	}

	switch s := resp.ExecutionReportStatus; s {
	case investpb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_FILL:
		return resp, nil

	case investpb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_REJECTED:
		return nil, ErrOrderRejected

	case investpb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_CANCELLED:
		return nil, ErrOrderCancelled

	default:
		return nil, fmt.Errorf("unexpected response status: %d", s)
	}
}

func adaptPbPostOrderResponse(resp *investpb.PostOrderResponse) *PlaceOrderResponse {
	return &PlaceOrderResponse{
		OrderID: resp.OrderId,
		ExecutedPrice: Quotation{
			Units: int(resp.ExecutedOrderPrice.Units),
			Nano:  int(resp.ExecutedOrderPrice.Units),
		},
	}
}

func adaptQuotationToPb(q Quotation) *investpb.Quotation {
	return &investpb.Quotation{
		Units: int64(q.Units),
		Nano:  int32(q.Nano),
	}
}
