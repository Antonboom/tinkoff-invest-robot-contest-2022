package main

import (
	"math/rand"
	"time"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type Simulator struct {
	investpb.UnimplementedInstrumentsServiceServer
	investpb.UnimplementedMarketDataStreamServiceServer
	investpb.UnimplementedSandboxServiceServer
}

func NewSimulator() *Simulator {
	return new(Simulator)
}
