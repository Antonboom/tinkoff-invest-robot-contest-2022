package tinkoffinvest

import (
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

type Client struct {
	token      string
	appName    string
	useSandbox bool

	instruments      investpb.InstrumentsServiceClient
	marketDataStream investpb.MarketDataStreamServiceClient
	orders           investpb.OrdersServiceClient

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
		marketDataStream: investpb.NewMarketDataStreamServiceClient(cc),
		orders:           investpb.NewOrdersServiceClient(cc),
		sandbox:          investpb.NewSandboxServiceClient(cc),
	}, nil
}

func (c *Client) auth(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx,
		"authorization", "Bearer "+c.token,
		"x-app-name", c.appName)
}
