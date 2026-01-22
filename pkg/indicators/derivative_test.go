package indicators_test

import (
	"testing"

	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/irfndi/goflux/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

func TestDerivativeIndicator(t *testing.T) {
	series := testutils.MockTimeSeries("1", "1", "2", "3", "5", "8", "13")
	indicator := DerivativeIndicator{
		indicators.Indicator: indicators.NewClosePriceIndicator(series),
	}

	t.Run("returns zero at index zero", func(t *testing.T) {
		assert.EqualValues(t, "0", indicator.Calculate(0).String())
	})

	t.Run("returns the derivative", func(t *testing.T) {
		assert.EqualValues(t, "0", indicator.Calculate(1).String())

		for i := 2; i < len(series.series.Candles); i++ {
			expected := series.series.Candles[i-2].ClosePrice

			assert.EqualValues(t, expected.String(), indicator.Calculate(i).String())
		}
	})
}
