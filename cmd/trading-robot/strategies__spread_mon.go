package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest"
	"github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/config"
)

func runSpreadMonitoring(
	ctx context.Context,
	account string,
	cfg config.SpreadMonitoringConfig,
	client *tinkoffinvest.Client,
) error {
	tools, err := client.GetTradeAvailableShares(ctx)
	if err != nil {
		return fmt.Errorf("get available instruments: %v", err)
	}

	filteredTools := make([]tinkoffinvest.Instrument, 0, len(tools))
	for _, t := range tools {
		logger := log.With().Str("figi", t.FIGI).Logger()

		orderBook, err := client.GetOrderBook(ctx, tinkoffinvest.OrderBookRequest{
			FIGI:  t.FIGI,
			Depth: cfg.Depth,
		})
		if err != nil {
			logger.Err(err).Msg("get order book")
			continue
		}

		spread := tinkoffinvest.Spread(*orderBook)

		logger.Debug().
			Str("name", t.Name).
			Float64("spread", spread).
			Msg("spread")

		if spread >= cfg.MinSpreadPercentage {
			filteredTools = append(filteredTools, t)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(800 * time.Millisecond):
		}
	}

	_ = json.NewEncoder(os.Stdout).Encode(filteredTools)

	return nil
}
