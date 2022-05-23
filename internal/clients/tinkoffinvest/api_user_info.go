package tinkoffinvest

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"

	investpb "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest/pb"
)

var ErrInvalidToken = errors.New("invalid token")

type UserInfo struct {
	PremStatus           bool
	QualStatus           bool
	QualifiedForWorkWith []string
	Tariff               string
}

func (c *Client) GetUserInfo(ctx context.Context) (*UserInfo, error) {
	resp, err := c.users.GetInfo(c.auth(ctx), new(investpb.GetInfoRequest))
	if err != nil {
		if isStatusError(err, codes.Unauthenticated, "") {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("grpc user info call: %v", err)
	}

	return &UserInfo{
		PremStatus:           resp.PremStatus,
		QualStatus:           resp.QualStatus,
		QualifiedForWorkWith: resp.QualifiedForWorkWith,
		Tariff:               resp.Tariff,
	}, nil
}
