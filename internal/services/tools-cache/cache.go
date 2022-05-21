package toolscache

import (
	"context"
	"fmt"
	"sync"

	"github.com/shopspring/decimal"

	"github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest"
)

type SharesProvider interface {
	GetShareByFIGI(ctx context.Context, figi tinkoffinvest.FIGI) (*tinkoffinvest.Instrument, error)
}

// Cache implements quite simple instruments cache.
type Cache struct {
	mu       *sync.Mutex
	tools    map[tinkoffinvest.FIGI]Tool
	provider SharesProvider
}

type Tool struct {
	FIGI         tinkoffinvest.FIGI
	StocksPerLot int
	MinPriceInc  decimal.Decimal
}

func New(p SharesProvider) *Cache {
	return &Cache{
		tools:    make(map[tinkoffinvest.FIGI]Tool),
		mu:       new(sync.Mutex),
		provider: p,
	}
}

func (c *Cache) Get(ctx context.Context, figi tinkoffinvest.FIGI) (Tool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if t, ok := c.tools[figi]; ok {
		return t, nil
	}

	share, err := c.provider.GetShareByFIGI(ctx, figi)
	if err != nil {
		return Tool{}, fmt.Errorf("get share by figi: %v", err)
	}

	minPriceInc, err := decimal.NewFromString(share.MinPriceIncrement)
	if err != nil {
		return Tool{}, fmt.Errorf("invalid share.MinPriceIncrement value %v: %v", share.MinPriceIncrement, err)
	}

	tool := Tool{
		FIGI:         share.FIGI,
		StocksPerLot: share.Lot,
		MinPriceInc:  minPriceInc,
	}
	c.tools[figi] = tool
	return tool, nil
}
