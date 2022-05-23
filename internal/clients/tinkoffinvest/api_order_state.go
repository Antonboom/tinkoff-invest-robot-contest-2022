package tinkoffinvest

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

const pollOrderStateInterval = 100 * time.Millisecond

var (
	ErrOrderWaitExecution = errors.New("order is waiting for execution")
	ErrOrderRejected      = errors.New("order rejected")
	ErrOrderCancelled     = errors.New("order cancelled by user")
)

// WaitForOrderExecution allows waiting order execution via simple polling with GetOrderState method.
// Useful because investpb.OrdersStreamServiceClient is not working in sandbox.
func (c *Client) WaitForOrderExecution(ctx context.Context, accountID AccountID, orderID OrderID) (decimal.Decimal, error) {
	for {
		select {
		case <-ctx.Done():
			return decimal.Zero, ctx.Err()

		case <-time.After(pollOrderStateInterval):
			price, err := c.GetOrderState(ctx, accountID, orderID)
			if err != nil {
				if errors.Is(err, ErrOrderWaitExecution) {
					continue
				}
				return decimal.Zero, err
			}
			return price, nil
		}
	}
}

// GetOrderState returns executed price if the order was executed and one of errors otherwise:
// - ErrOrderWaitExecution
// - ErrOrderRejected
// - ErrOrderCancelled.
func (c *Client) GetOrderState(ctx context.Context, accountID AccountID, orderID OrderID) (decimal.Decimal, error) {
	ctx = c.auth(ctx)

	req := &investpb.GetOrderStateRequest{
		AccountId: accountID.S(),
		OrderId:   orderID.S(),
	}

	var (
		resp *investpb.OrderState
		err  error
	)
	if c.useSandbox {
		resp, err = c.sandbox.GetSandboxOrderState(ctx, req)
	} else {
		resp, err = c.orders.GetOrderState(ctx, req)
	}
	if err != nil {
		return decimal.Zero, fmt.Errorf("grpc get order state call: %v", err)
	}

	switch s := resp.ExecutionReportStatus; s {
	case investpb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_FILL:
		return newDecimal(
			resp.ExecutedOrderPrice.Units,
			resp.ExecutedOrderPrice.Nano,
		), nil

	case investpb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_NEW,
		investpb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_PARTIALLYFILL:
		return decimal.Zero, ErrOrderWaitExecution

	case investpb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_REJECTED:
		return decimal.Zero, ErrOrderRejected

	case investpb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_CANCELLED:
		return decimal.Zero, ErrOrderCancelled

	default:
		return decimal.Zero, fmt.Errorf("unexpected order status: %d", s)
	}
}
