package backtest

import "github.com/irfndi/goflux/pkg/decimal"

// equityMark tracks unrealized profit without recalculating every position
// from its entry price on every bar.
type equityMark struct {
	value       decimal.Decimal
	netQuantity decimal.Decimal
	lastPrice   decimal.Decimal
	hasPrice    bool
}

func (m *equityMark) advance(price decimal.Decimal) {
	if !m.hasPrice {
		m.lastPrice = price
		m.hasPrice = true
		return
	}
	if !m.netQuantity.IsZero() {
		m.value = m.value.Add(price.Sub(m.lastPrice).Mul(m.netQuantity))
	}
	m.lastPrice = price
}

func (m *equityMark) add(direction string, entryPrice, quantity, currentPrice decimal.Decimal) {
	if quantity.IsZero() {
		return
	}
	if !m.hasPrice {
		m.lastPrice = currentPrice
		m.hasPrice = true
	}
	if direction == "short" {
		m.netQuantity = m.netQuantity.Sub(quantity)
		m.value = m.value.Add(entryPrice.Sub(currentPrice).Mul(quantity))
		return
	}
	m.netQuantity = m.netQuantity.Add(quantity)
	m.value = m.value.Add(currentPrice.Sub(entryPrice).Mul(quantity))
}

func (m *equityMark) remove(direction string, entryPrice, quantity, currentPrice decimal.Decimal) {
	if quantity.IsZero() {
		return
	}
	if direction == "short" {
		m.netQuantity = m.netQuantity.Add(quantity)
		m.value = m.value.Sub(entryPrice.Sub(currentPrice).Mul(quantity))
		return
	}
	m.netQuantity = m.netQuantity.Sub(quantity)
	m.value = m.value.Sub(currentPrice.Sub(entryPrice).Mul(quantity))
}
