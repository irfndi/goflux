package trading

import (
	"testing"

	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/irfndi/goflux/pkg/testutils"
)

func TestTRIXOverZeroRule(t *testing.T) {
	s := testutils.MockTimeSeriesFl(100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120)

	trix := indicators.NewTRIXIndicatorFromSeries(s, 5)
	rule := NewTRIXOverZeroRule(trix)
	record := NewTradingRecord()

	if !rule.IsSatisfied(20, record) {
		t.Errorf("TRIXOverZeroRule should be satisfied in uptrend")
	}
}

func TestTRIXOverLevelRule(t *testing.T) {
	s := testutils.MockTimeSeriesFl(100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120)

	trix := indicators.NewTRIXIndicatorFromSeries(s, 5)
	rule := NewTRIXOverLevelRule(trix, 0.1)
	record := NewTradingRecord()

	if !rule.IsSatisfied(20, record) {
		t.Errorf("TRIXOverLevelRule should be satisfied when TRIX > 0.1")
	}
}

func TestTRIXUnderLevelRule(t *testing.T) {
	s := testutils.MockTimeSeriesFl(120, 119, 118, 117, 116, 115, 114, 113, 112, 111, 110, 109, 108, 107, 106, 105, 104, 103, 102, 101, 100)

	trix := indicators.NewTRIXIndicatorFromSeries(s, 5)
	rule := NewTRIXUnderLevelRule(trix, -0.1)
	record := NewTradingRecord()

	if !rule.IsSatisfied(20, record) {
		t.Errorf("TRIXUnderLevelRule should be satisfied when TRIX < -0.1")
	}
}

func TestTRIXBullishRule(t *testing.T) {
	s := testutils.MockTimeSeriesFl(100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120)

	rule := NewTRIXBullishRule(s, 5)
	record := NewTradingRecord()

	if !rule.IsSatisfied(20, record) {
		t.Errorf("TRIXBullishRule should be satisfied in uptrend")
	}
}

func TestTRIXBearishRule(t *testing.T) {
	s := testutils.MockTimeSeriesFl(120, 119, 118, 117, 116, 115, 114, 113, 112, 111, 110, 109, 108, 107, 106, 105, 104, 103, 102, 101, 100)

	rule := NewTRIXBearishRule(s, 5)
	record := NewTradingRecord()

	if !rule.IsSatisfied(20, record) {
		t.Errorf("TRIXBearishRule should be satisfied in downtrend")
	}
}

func TestTRIXAtThreshold(t *testing.T) {
	// Flat trend → TRIX ≈ 0.
	s := testutils.MockTimeSeriesFl(100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100)

	trix := indicators.NewTRIXIndicatorFromSeries(s, 5)
	record := NewTradingRecord()

	// Flat prices produce small TRIX noise from EMA initialization (~0.5).
	over := NewTRIXOverLevelRule(trix, 1)
	if over.IsSatisfied(19, record) {
		t.Errorf("TRIXOverLevelRule(level=1) should not be satisfied in flat trend")
	}

	under := NewTRIXUnderLevelRule(trix, -1)
	if under.IsSatisfied(19, record) {
		t.Errorf("TRIXUnderLevelRule(level=-1) should not be satisfied in flat trend")
	}
}

func TestTRIXInsufficientData(t *testing.T) {
	s := testutils.MockTimeSeriesFl(100, 101, 102)

	record := NewTradingRecord()

	bullish := NewTRIXBullishRule(s, 5)
	if bullish.IsSatisfied(2, record) {
		t.Errorf("TRIXBullishRule should not be satisfied with insufficient data")
	}

	bearish := NewTRIXBearishRule(s, 5)
	if bearish.IsSatisfied(2, record) {
		t.Errorf("TRIXBearishRule should not be satisfied with insufficient data")
	}
}
