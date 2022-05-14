package main

import (
	"context"
	"crypto/tls"
	"flag"
	stdlog "log"
	"os/signal"
	"strings"
	"syscall"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

	lvl, err := zerolog.ParseLevel(strings.ToLower(cfg.Log.Level))
	mustNil(err)
	zerolog.SetGlobalLevel(lvl)

	log.Info().Msg("connect to tinkoff invest api")
	conn, err := grpc.DialContext(ctx, cfg.Clients.TinkoffInvest.Address,
		grpc.WithBlock(),
		grpc.WithUserAgent(cfg.Clients.TinkoffInvest.AppName),
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})), //nolint:gosec
	)
	mustNil(err)

	tInvest, err := tinkoffinvest.NewClient(conn, cfg.Clients.TinkoffInvest.Token, cfg.Clients.TinkoffInvest.AppName)
	mustNil(err)

	books, err := tInvest.SubscribeForOrderBookChanges(ctx, []tinkoffinvest.OrderBookRequest{
		{
			Instrument: "BBG00YHVQ768",
			Depth:      10,
		},
	})
	mustNil(err)

	log.Info().Msg("listen Sberbank order book")
	for b := range books {
		log.Info().Msgf("%v", b)
	}
}

func mustNil(err error) {
	if err != nil {
		stdlog.Panic(err)
	}
}
