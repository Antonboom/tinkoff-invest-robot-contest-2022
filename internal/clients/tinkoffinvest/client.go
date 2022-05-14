package tinkoffinvest

import (
	"google.golang.org/grpc"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

type Client struct {
	instruments investpb.InstrumentsServiceClient
}

func NewClient(cc grpc.ClientConnInterface) *Client {
	return &Client{
		instruments: investpb.NewInstrumentsServiceClient(cc),
	}
}
