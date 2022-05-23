package main

import (
	stdlog "log"
	"net"

	"google.golang.org/grpc"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

const addr = ":7171"

func main() {
	srv := grpc.NewServer()
	sim := NewSimulator()

	investpb.RegisterInstrumentsServiceServer(srv, sim)
	investpb.RegisterMarketDataStreamServiceServer(srv, sim)
	investpb.RegisterSandboxServiceServer(srv, sim)

	lsn, err := net.Listen("tcp", addr)
	mustNil(err)

	stdlog.Printf("start grpc server at %q", addr)
	mustNil(srv.Serve(lsn))
}

func mustNil(err error) {
	if err != nil {
		stdlog.Panic(err)
	}
}
