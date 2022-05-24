package portfoliowatcher

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest"
)

//go:generate mockgen -source=$GOFILE -destination=mocks/watcher_generated.go -package portfoliowatchermocks PortfolioDataProvider

const defaultInterval = 5 * time.Second

type l = prometheus.Labels

type PortfolioDataProvider interface {
	GetBalance(ctx context.Context, accountID tinkoffinvest.AccountID) (decimal.Decimal, error)
	GetPortfolio(ctx context.Context, accountID tinkoffinvest.AccountID) (*tinkoffinvest.Portfolio, error)
}

type Watcher struct {
	interval    time.Duration
	account     tinkoffinvest.AccountID
	prevBalance decimal.Decimal
	provider    PortfolioDataProvider
}

func New(interval time.Duration, accountID tinkoffinvest.AccountID, provider PortfolioDataProvider) *Watcher {
	if interval <= 0 {
		interval = defaultInterval
	}
	return &Watcher{
		interval:    interval,
		account:     accountID,
		prevBalance: decimal.Zero,
		provider:    provider,
	}
}

func (w *Watcher) Run(ctx context.Context) error {
	if err := w.fetchAndSetAccountInfo(ctx); err != nil {
		log.Err(err).Msg("initial account info fetch")
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-time.After(w.interval):
			if err := w.fetchAndSetAccountInfo(ctx); err != nil {
				log.Err(err).Msg("periodic account info fetch")
			}
		}
	}
}

func (w *Watcher) fetchAndSetAccountInfo(ctx context.Context) error {
	balance, err := w.provider.GetBalance(ctx, w.account)
	if err != nil {
		return fmt.Errorf("get balance: %v", err)
	}

	portfolio, err := w.provider.GetPortfolio(ctx, w.account)
	if err != nil {
		return fmt.Errorf("get portfolio: %v", err)
	}

	if !balance.Equal(w.prevBalance) {
		w.prevBalance = balance

		log.Info().
			Str("service", "portfolio-watcher").
			Str("account", w.account.S()).
			Float64("balance", balance.InexactFloat64()).
			Msg("new account balance")
	}

	currentBalance.With(l{"account_number": w.account.S()}).Set(balance.InexactFloat64())
	sharesTotalPrice.With(l{"account_number": w.account.S()}).Set(portfolio.TotalSharesPrice.InexactFloat64())

	for _, share := range portfolio.Shares {
		shareQuantity.With(l{"account_number": w.account.S(), "figi": share.FIGI.S()}).Set(float64(share.Quantity))
		shareAvgPrice.With(l{"account_number": w.account.S(), "figi": share.FIGI.S()}).Set(share.AvgPrice.InexactFloat64())
	}

	return nil
}
