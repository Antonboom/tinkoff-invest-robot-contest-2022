package tinkoffinvest

import (
	"strconv"
)

const _10e9 = 1_000_000_000

type Quotation struct {
	Units int
	Nano  int
}

func (q Quotation) String() string {
	return strconv.Itoa(q.Units) + "." + strconv.Itoa(q.Nano/_10e9)
}

func (q Quotation) Sub(q2 Quotation) Quotation {
	return q.toNanos().sub(q2.toNanos()).toQuotation()
}

func (q Quotation) Mul(v float64) Quotation {
	return q.toNanos().mul(v).toQuotation()
}

func (q Quotation) Div(q2 Quotation) float64 {
	return q.toNanos().asFloat() / q2.toNanos().asFloat()
}

func (q Quotation) Greater(q2 Quotation) bool {
	return q.Units > q2.Units || (q.Units == q2.Units && q.Nano > q2.Nano)
}

func (q Quotation) Less(q2 Quotation) bool {
	return q.Units < q2.Units || (q.Units == q2.Units && q.Nano < q2.Nano)
}

func (q Quotation) toNanos() nanos {
	return nanos(q.Units*_10e9*100 + q.Nano) // Only EUR/USD/RUB and similar currencies are supported.
}

type nanos int

func (n nanos) sub(n2 nanos) nanos {
	return n - n2
}

func (n nanos) mul(v float64) nanos {
	return nanos(n.asFloat() * v)
}

func (n nanos) asFloat() float64 {
	return float64(n)
}

func (n nanos) toQuotation() Quotation {
	return Quotation{
		Units: int(n) / _10e9 / 100,
		Nano:  ((int(n) / _10e9) % 100) * _10e9,
	}
}
