package goflux

import (
	"testing"

	"github.com/irfndi/goflux/pkg/decimal"
)

func TestLegacyFactoryVariablesRemainAssignable(t *testing.T) {
	original := NewDecimal
	defer func() { NewDecimal = original }()

	NewDecimal = func(float64) decimal.Decimal { return decimal.ONE }
	if got := NewDecimal(42); !got.EQ(decimal.ONE) {
		t.Fatalf("assigned legacy NewDecimal factory returned %s", got)
	}
}
