package tinkoffinvest

import (
	"context"
	"errors"
	"fmt"
	"time"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

const pollOrderStateInterval = 100 * time.Millisecond

var (
	ErrOrderWaitExecution = errors.New("order is waiting for execution")
	ErrOrderRejected      = errors.New("order rejected")
	ErrOrderCancelled     = errors.New("order cancelled by user")
)

func (c *Client) WaitForOrderExecution(ctx context.Context, accountID AccountID, orderID OrderID) (*Quotation, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()

		case <-time.After(pollOrderStateInterval):
			price, err := c.GetOrderState(ctx, accountID, orderID)
			if err != nil {
				if errors.Is(err, ErrOrderWaitExecution) {
					continue
				}
				return nil, err
			}
			return price, nil
		}
	}
}

// GetOrderState returns executed price if the order was executed and an error otherwise..
func (c *Client) GetOrderState(ctx context.Context, accountID AccountID, orderID OrderID) (*Quotation, error) {
	ctx = c.auth(ctx)

	req := &investpb.GetOrderStateRequest{
		AccountId: string(accountID),
		OrderId:   string(orderID),
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
		return nil, fmt.Errorf("grpc get order state call: %v", err)
	}

	switch s := resp.ExecutionReportStatus; s {
	case investpb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_FILL:
		return &Quotation{
			Units: int(resp.ExecutedOrderPrice.Units),
			Nano:  int(resp.ExecutedOrderPrice.Nano),
		}, nil

	case investpb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_NEW:
		return nil, ErrOrderWaitExecution

	case investpb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_REJECTED:
		return nil, ErrOrderRejected

	case investpb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_CANCELLED:
		return nil, ErrOrderCancelled

	default:
		return nil, fmt.Errorf("unexpected order status: %d", s)
	}
}
