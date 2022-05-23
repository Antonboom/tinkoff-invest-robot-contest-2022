package tinkoffinvest

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

const (
	codeNotEnoughAssetsForMarginTrade = "30042"
)

type Client struct {
	token      string
	appName    string
	useSandbox bool

	instruments      investpb.InstrumentsServiceClient
	marketData       investpb.MarketDataServiceClient
	marketDataStream investpb.MarketDataStreamServiceClient
	operations       investpb.OperationsServiceClient
	orders           investpb.OrdersServiceClient
	users            investpb.UsersServiceClient

	sandbox investpb.SandboxServiceClient
}

func NewClient(cc grpc.ClientConnInterface, token string, appName string, useSandbox bool) (*Client, error) {
	if cc == nil {
		return nil, errors.New("uninitialized grpc connection")
	}
	if token == "" {
		return nil, errors.New("api token must be defined")
	}
	if appName == "" {
		return nil, errors.New("application name must be defined")
	}

	return &Client{
		token:            token,
		appName:          appName,
		useSandbox:       useSandbox,
		instruments:      investpb.NewInstrumentsServiceClient(cc),
		marketData:       investpb.NewMarketDataServiceClient(cc),
		marketDataStream: investpb.NewMarketDataStreamServiceClient(cc),
		operations:       investpb.NewOperationsServiceClient(cc),
		orders:           investpb.NewOrdersServiceClient(cc),
		users:            investpb.NewUsersServiceClient(cc),
		sandbox:          investpb.NewSandboxServiceClient(cc),
	}, nil
}

func (c *Client) auth(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx,
		"authorization", "Bearer "+c.token,
		"x-app-name", c.appName)
}

func isStatusError(err error, typpe codes.Code, code string) bool {
	s, ok := status.FromError(err)
	if !ok {
		return false
	}
	if s.Code() != typpe {
		return false
	}
	if code != "" {
		return s.Message() == code
	}
	return true
}
