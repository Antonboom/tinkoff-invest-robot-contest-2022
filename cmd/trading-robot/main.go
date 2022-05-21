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
	toolscache "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/services/tools-cache"
	bullsbearsmon "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/strategies/bulls-and-bears-mon"
	spreadparasite "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/strategies/spread-parasite"
)

var configPath = flag.String("config", "configs/config.toml", "Path to config file")

func init() {
	flag.Parse()
}

type Strategy interface {
	Name() string
	Run(ctx context.Context) error
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

	tInvestClient, err := tinkoffinvest.NewClient(
		conn,
		cfg.Clients.TinkoffInvest.Token,
		cfg.Clients.TinkoffInvest.AppName,
		cfg.Clients.TinkoffInvest.UseSandbox,
	)
	mustNil(err)

	toolsCache := toolscache.New(tInvestClient)

	if !cfg.Clients.TinkoffInvest.UseSandbox {
		_, err = tInvestClient.GetUserInfo(ctx)
		if errors.Is(err, tinkoffinvest.ErrInvalidToken) {
			stdlog.Panic("unauthenticated: invalid clients.tinkfoff_invest.token")
			return
		}
		mustNil(err)
	}

	var wg Waiter
	errCh := make(chan error, 3)

	if cfg.Metrics.Enabled {
		wg.Go(func() { errCh <- runMetrics(ctx, cfg.Metrics.Addr) })
	}

	var strategies []Strategy

	if bbMonCfg := cfg.Strategies.BullsAndBearsMonitoring; bbMonCfg.Enabled {
		toolConfs := make([]bullsbearsmon.ToolConfig, len(bbMonCfg.Instruments))
		for i, ins := range bbMonCfg.Instruments {
			toolConfs[i] = bullsbearsmon.ToolConfig{
				FIGI:             tinkoffinvest.FIGI(ins.FIGI),
				Depth:            ins.Depth,
				DominanceRatio:   ins.DominanceRatio,
				ProfitPercentage: ins.ProfitPercentage,
			}
		}

		strategy, err := bullsbearsmon.New(
			tinkoffinvest.AccountID(cfg.Account.Number),
			bbMonCfg.IgnoreInconsistent,
			toolConfs,
			tInvestClient,
			toolsCache,
		)
		mustNil(err)

		strategies = append(strategies, strategy)
	}

	if spCfg := cfg.Strategies.SpreadParasite; spCfg.Enabled {
		strategy, err := spreadparasite.New(
			tinkoffinvest.AccountID(cfg.Account.Number),
			spCfg.IgnoreInconsistent,
			spCfg.Depth,
			spCfg.MinSpreadPercentage,
			spCfg.MaxTools,
			tInvestClient,
			toolsCache,
		)
		mustNil(err)

		strategies = append(strategies, strategy)
	}

	if len(strategies) == 0 {
		log.Warn().Msg("no strategies enabled")
		cancel()
	}
	for _, s := range strategies {
		s := s
		wg.Go(func() { errCh <- s.Run(ctx) })
	}

	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil {
			log.Err(err).Msg("error on startup")
			cancel()
		}
	}

	log.Info().Msg("shutdown")
	wg.Wait()
}

func mustNil(err error) {
	if err != nil {
		stdlog.Panic(err)
	}
}
