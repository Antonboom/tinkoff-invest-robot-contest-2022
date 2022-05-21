package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	stdlog "log"
	"os"
	"os/signal"
	"sort"
	"syscall"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest"
	"github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/config"
)

var configPath = flag.String("config", "configs/config.toml", "Path to config file")

func init() {
	flag.Parse()
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Parse(*configPath)
	mustNil(err)
	mustNil(validator.New().Struct(cfg))

	conn, err := grpc.DialContext(ctx, cfg.Clients.TinkoffInvest.Address,
		grpc.WithBlock(),
		grpc.WithUserAgent(cfg.Clients.TinkoffInvest.AppName),
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})), //nolint:gosec
	)
	mustNil(err)

	tInvest, err := tinkoffinvest.NewClient(
		conn,
		cfg.Clients.TinkoffInvest.Token,
		cfg.Clients.TinkoffInvest.AppName,
		true,
	)
	mustNil(err)

	instruments, err := tInvest.GetTradeAvailableShares(ctx)
	mustNil(err)

	sort.Slice(instruments, func(i, j int) bool {
		return instruments[i].Name < instruments[j].Name
	})
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	mustNil(enc.Encode(instruments))
}

func mustNil(err error) {
	if err != nil {
		stdlog.Panic(err)
	}
}
