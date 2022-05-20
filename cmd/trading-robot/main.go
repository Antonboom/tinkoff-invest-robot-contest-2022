package main

import (
	"context"
	"crypto/tls"
	"errors"
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

	tInvest, err := tinkoffinvest.NewClient(
		conn,
		cfg.Clients.TinkoffInvest.Token,
		cfg.Clients.TinkoffInvest.AppName,
		cfg.Clients.TinkoffInvest.UseSandbox,
	)
	mustNil(err)

	if !cfg.Clients.TinkoffInvest.UseSandbox {
		_, err = tInvest.GetUserInfo(ctx)
		if errors.Is(err, tinkoffinvest.ErrInvalidToken) {
			stdlog.Panic("unauthenticated: invalid clients.tinkfoff_invest.token")
			return
		}
		mustNil(err)
	}

	if cfg.Metrics.Enabled {
		runMetrics(cfg.Metrics.Addr)
	}

	switch {
	case cfg.Strategies.BullsAndBearsMonitoring.Enabled:
		strategyCfg := cfg.Strategies.BullsAndBearsMonitoring
		if err := runBullsAndBearsMonitoring(ctx, cfg.Account.Number, strategyCfg, tInvest); err != nil {
			log.Err(err).Msg("cannot run bulls and bears monitoring strategy")
		}

	case cfg.Strategies.SpreadMonitoring.Enabled:
		strategyCfg := cfg.Strategies.SpreadMonitoring
		if err := runSpreadMonitoring(ctx, cfg.Account.Number, strategyCfg, tInvest); err != nil {
			log.Err(err).Msg("cannot run spread monitoring strategy")
		}

	default:
		log.Warn().Msg("no strategies enabled: exit")
		cancel()
	}

	<-ctx.Done()
}

func mustNil(err error) {
	if err != nil {
		stdlog.Panic(err)
	}
}
