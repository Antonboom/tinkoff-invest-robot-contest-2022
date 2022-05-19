package tinkoffinvest

import (
	"strconv"
)

type Quotation struct {
	Units int
	Nano  int
}

func (q Quotation) String() string {
	return strconv.Itoa(q.Units) + "." + strconv.Itoa(q.Nano)
}

func (q Quotation) Mul(v float64) Quotation {
	nanos := q.Units*100 + q.Nano // Only EUR/USD/RUB and similar currencies are supported.
	r := int(float64(nanos) * v)
	return Quotation{
		Units: r / 100,
		Nano:  r % 100,
	}
}

func (q Quotation) Greater(q2 Quotation) bool {
	return q.Units > q2.Units || (q.Units == q2.Units && q.Nano > q2.Nano)
}

func (q Quotation) Less(q2 Quotation) bool {
	return q.Units < q2.Units || (q.Units == q2.Units && q.Nano < q2.Nano)
}
