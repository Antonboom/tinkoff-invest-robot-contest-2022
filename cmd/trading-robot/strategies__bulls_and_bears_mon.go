package main

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest"
	"github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/config"
	bullsbearsmon "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/strategies/bulls-and-bears-mon"
)

func runBullsAndBearsMonitoring(
	ctx context.Context,
	account string,
	cfg config.BullsAndBearsMonitoringConfig,
	client *tinkoffinvest.Client,
) error {
	reqs := make([]tinkoffinvest.OrderBookRequest, len(cfg.Instruments))
	toolConfs := make([]bullsbearsmon.ToolConfig, len(cfg.Instruments))
	for i, ins := range cfg.Instruments {
		reqs[i] = tinkoffinvest.OrderBookRequest{
			FIGI:  ins.FIGI,
			Depth: ins.Depth,
		}
		toolConfs[i] = bullsbearsmon.ToolConfig{
			FIGI:           ins.FIGI,
			DominanceRatio: ins.DominanceRatio,
		}
	}

	strategy, err := bullsbearsmon.New(account, cfg.IgnoreInconsistent, toolConfs, client)
	if err != nil {
		return fmt.Errorf("build strategy: %v", err)
	}

	changes, err := client.SubscribeForOrderBookChanges(ctx, reqs)
	if err != nil {
		return fmt.Errorf("subscribe for order book changes: %v", err)
	}

	go func() {
		logger := log.With().Str("strategy", strategy.Name()).Logger()

		for {
			select {
			case <-time.After(5 * time.Second):
				logger.Debug().Msg("no order book changes due to period")

			case change, ok := <-changes:
				if !ok {
					return
				}

				func() {
					ctx, cancel := context.WithTimeout(ctx, time.Second)
					defer cancel()

					if err := strategy.Apply(ctx, change); err != nil {
						logger.Err(err).Msg("cannot apply order book change")
					}
				}()
			}
		}
	}()
	return nil
}
