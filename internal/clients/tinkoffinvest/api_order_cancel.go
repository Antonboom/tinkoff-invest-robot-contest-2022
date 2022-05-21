package tinkoffinvest

import (
	"context"
	"fmt"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

func (c *Client) CancelOrder(ctx context.Context, orderID OrderID) error {
	ctx = c.auth(ctx)

	req := &investpb.CancelOrderRequest{
		OrderId: orderID.S(),
	}

	var err error
	if c.useSandbox {
		_, err = c.sandbox.CancelSandboxOrder(ctx, req)
	} else {
		_, err = c.orders.CancelOrder(ctx, req)
	}
	if err != nil {
		return fmt.Errorf("grpc canel order call: %v", err)
	}
	return nil
}
