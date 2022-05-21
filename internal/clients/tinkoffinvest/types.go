package tinkoffinvest

// For compile-time restrictions.

type FIGI string          //
func (id FIGI) S() string { return string(id) }

type AccountID string          //
func (id AccountID) S() string { return string(id) }

type OrderID string          //
func (id OrderID) S() string { return string(id) }
