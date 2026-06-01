package trading

import (
	"testing"

	"github.com/irfndi/goflux/pkg/indicators"
	"github.com/irfndi/goflux/pkg/testutils"
)

func TestTRIXOverZeroRule(t *testing.T) {
	s := testutils.MockTimeSeriesFl(100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130)

	trix := indicators.NewTRIXIndicatorFromSeries(s, 5)
	rule := NewTRIXOverLevelRule(trix, 0)
	record := NewTradingRecord()

	if !rule.IsSatisfied(30, record) {
		t.Errorf("TRIXOverZeroRule should be satisfied in uptrend")
	}
}

func TestTRIXOverLevelRule(t *testing.T) {
	s := testutils.MockTimeSeriesFl(100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130)

	trix := indicators.NewTRIXIndicatorFromSeries(s, 5)
	rule := NewTRIXOverLevelRule(trix, 0.1)
	record := NewTradingRecord()

	if !rule.IsSatisfied(30, record) {
		t.Errorf("TRIXOverLevelRule should be satisfied when TRIX > 0.1")
	}
}

func TestTRIXUnderLevelRule(t *testing.T) {
	s := testutils.MockTimeSeriesFl(130, 129, 128, 127, 126, 125, 124, 123, 122, 121, 120, 119, 118, 117, 116, 115, 114, 113, 112, 111, 110, 109, 108, 107, 106, 105, 104, 103, 102, 101, 100)

	trix := indicators.NewTRIXIndicatorFromSeries(s, 5)
	rule := NewTRIXUnderLevelRule(trix, -0.1)
	record := NewTradingRecord()

	if !rule.IsSatisfied(30, record) {
		t.Errorf("TRIXUnderLevelRule should be satisfied when TRIX < -0.1")
	}
}

func TestTRIXBullishRule(t *testing.T) {
	s := testutils.MockTimeSeriesFl(100, 101, 102, 103, 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127, 128, 129, 130)

	rule := NewTRIXBullishRule(s, 5)
	record := NewTradingRecord()

	if !rule.IsSatisfied(30, record) {
		t.Errorf("TRIXBullishRule should be satisfied in uptrend")
	}
}

func TestTRIXBearishRule(t *testing.T) {
	s := testutils.MockTimeSeriesFl(130, 129, 128, 127, 126, 125, 124, 123, 122, 121, 120, 119, 118, 117, 116, 115, 114, 113, 112, 111, 110, 109, 108, 107, 106, 105, 104, 103, 102, 101, 100)

	rule := NewTRIXBearishRule(s, 5)
	record := NewTradingRecord()

	if !rule.IsSatisfied(30, record) {
		t.Errorf("TRIXBearishRule should be satisfied in downtrend")
	}
}

func TestTRIXAtThreshold(t *testing.T) {
	// Flat trend → TRIX ≈ 0 (need ≥6*window data for warmup)
	s := testutils.MockTimeSeriesFl(100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100, 100)

	trix := indicators.NewTRIXIndicatorFromSeries(s, 5)
	record := NewTradingRecord()

	over := NewTRIXOverLevelRule(trix, 1)
	if over.IsSatisfied(35, record) {
		t.Errorf("TRIXOverLevelRule(level=1) should not be satisfied in flat trend")
	}

	under := NewTRIXUnderLevelRule(trix, -1)
	if under.IsSatisfied(35, record) {
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
