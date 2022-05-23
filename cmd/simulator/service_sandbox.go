package main

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

func (s *Simulator) PostSandboxOrder(context.Context, *investpb.PostOrderRequest) (*investpb.PostOrderResponse, error) {
	return &investpb.PostOrderResponse{
		OrderId: uuid.NewString(),
	}, nil
}

func (s *Simulator) CancelSandboxOrder(context.Context, *investpb.CancelOrderRequest) (*investpb.CancelOrderResponse, error) {
	return &investpb.CancelOrderResponse{
		Time: timestamppb.Now(),
	}, nil
}

func (s *Simulator) GetSandboxOrderState(ctx context.Context, req *investpb.GetOrderStateRequest) (*investpb.OrderState, error) {
	return &investpb.OrderState{
		OrderId:               req.OrderId,
		ExecutionReportStatus: investpb.OrderExecutionReportStatus_EXECUTION_REPORT_STATUS_FILL,
		ExecutedOrderPrice:    &investpb.MoneyValue{Currency: "rub", Units: 122, Nano: 330000000},
	}, nil
}

func (s *Simulator) GetSandboxPortfolio(context.Context, *investpb.PortfolioRequest) (*investpb.PortfolioResponse, error) {
	return &investpb.PortfolioResponse{
		TotalAmountShares: &investpb.MoneyValue{Currency: "rub", Units: 1000, Nano: 0},
		Positions:         nil,
	}, nil
}

func (s *Simulator) SandboxPayIn(context.Context, *investpb.SandboxPayInRequest) (*investpb.SandboxPayInResponse, error) {
	return &investpb.SandboxPayInResponse{
		Balance: &investpb.MoneyValue{Currency: "rub", Units: 1000, Nano: 0},
	}, nil
}
