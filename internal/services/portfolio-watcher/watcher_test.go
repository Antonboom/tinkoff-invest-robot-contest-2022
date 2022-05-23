package portfoliowatcher_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"

	"github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest"
	portfoliowatcher "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/services/portfolio-watcher"
	portfoliowatchermocks "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/services/portfolio-watcher/mocks"
)

const accountID = tinkoffinvest.AccountID("account-yyy")

func TestWatcher(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	provider := portfoliowatchermocks.NewMockPortfolioDataProvider(ctrl)
	w := portfoliowatcher.New(accountID, provider)

	ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
	defer cancel()

	provider.EXPECT().GetBalance(gomock.Any(), accountID).Return(decimal.RequireFromString("100000"), nil).Times(2)
	provider.EXPECT().GetPortfolio(gomock.Any(), accountID).Return(&tinkoffinvest.Portfolio{
		TotalSharesPrice: decimal.RequireFromString("99000"),
		Shares: []tinkoffinvest.PortfolioPosition{
			{
				FIGI:     "BBG004S68473",
				Quantity: 10,
				AvgPrice: decimal.RequireFromString("1000"),
			},
			{
				FIGI:     "BBG000NL6ZD9",
				Quantity: 23,
				AvgPrice: decimal.RequireFromString("2500"),
			},
		},
	}, nil).Times(2)

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = w.Run(ctx)
	}()

	<-done
}
