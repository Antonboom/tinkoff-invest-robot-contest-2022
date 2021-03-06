// Code generated by MockGen. DO NOT EDIT.
// Source: watcher.go

// Package portfoliowatchermocks is a generated GoMock package.
package portfoliowatchermocks

import (
	context "context"
	reflect "reflect"

	tinkoffinvest "github.com/Antonboom/tinkoff-invest-robot-contest-2022/internal/clients/tinkoffinvest"
	gomock "github.com/golang/mock/gomock"
	decimal "github.com/shopspring/decimal"
)

// MockPortfolioDataProvider is a mock of PortfolioDataProvider interface.
type MockPortfolioDataProvider struct {
	ctrl     *gomock.Controller
	recorder *MockPortfolioDataProviderMockRecorder
}

// MockPortfolioDataProviderMockRecorder is the mock recorder for MockPortfolioDataProvider.
type MockPortfolioDataProviderMockRecorder struct {
	mock *MockPortfolioDataProvider
}

// NewMockPortfolioDataProvider creates a new mock instance.
func NewMockPortfolioDataProvider(ctrl *gomock.Controller) *MockPortfolioDataProvider {
	mock := &MockPortfolioDataProvider{ctrl: ctrl}
	mock.recorder = &MockPortfolioDataProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPortfolioDataProvider) EXPECT() *MockPortfolioDataProviderMockRecorder {
	return m.recorder
}

// GetBalance mocks base method.
func (m *MockPortfolioDataProvider) GetBalance(ctx context.Context, accountID tinkoffinvest.AccountID) (decimal.Decimal, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBalance", ctx, accountID)
	ret0, _ := ret[0].(decimal.Decimal)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBalance indicates an expected call of GetBalance.
func (mr *MockPortfolioDataProviderMockRecorder) GetBalance(ctx, accountID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBalance", reflect.TypeOf((*MockPortfolioDataProvider)(nil).GetBalance), ctx, accountID)
}

// GetPortfolio mocks base method.
func (m *MockPortfolioDataProvider) GetPortfolio(ctx context.Context, accountID tinkoffinvest.AccountID) (*tinkoffinvest.Portfolio, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPortfolio", ctx, accountID)
	ret0, _ := ret[0].(*tinkoffinvest.Portfolio)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPortfolio indicates an expected call of GetPortfolio.
func (mr *MockPortfolioDataProviderMockRecorder) GetPortfolio(ctx, accountID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPortfolio", reflect.TypeOf((*MockPortfolioDataProvider)(nil).GetPortfolio), ctx, accountID)
}
