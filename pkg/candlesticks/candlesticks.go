package candlesticks

import (
	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/series"
)

type Candle struct {
	Open  decimal.Decimal
	High  decimal.Decimal
	Low   decimal.Decimal
	Close decimal.Decimal
}

func NewCandle(c *series.Candle) Candle {
	if c == nil {
		return Candle{}
	}
	return Candle{
		Open:  c.OpenPrice,
		High:  c.MaxPrice,
		Low:   c.MinPrice,
		Close: c.ClosePrice,
	}
}

type Pattern int

const (
	None Pattern = iota
	Doji
	DragonflyDoji
	GravestoneDoji
	Hammer
	InvertedHammer
	HangingMan
	ShootingStar
	BullishEngulfing
	BearishEngulfing
	BullishHarami
	BearishHarami
	MorningStar
	EveningStar
	ThreeWhiteSoldiers
	ThreeBlackCrows
	SpinningTop
	Marubozu
	DojiStar
	PiercingLine
	DarkCloudCover
)

type PatternDetector struct {
	series *series.TimeSeries
}

func NewPatternDetector(s *series.TimeSeries) *PatternDetector {
	return &PatternDetector{series: s}
}

func (pd *PatternDetector) GetCandle(index int) Candle {
	if index < 0 || index >= pd.Length() {
		return Candle{}
	}
	candle := pd.series.Candles[index]
	if candle == nil {
		return Candle{}
	}
	return NewCandle(candle)
}

func (pd *PatternDetector) Length() int {
	return len(pd.series.Candles)
}

func (pd *PatternDetector) Detect(index int) Pattern {
	if index < 0 || index >= pd.Length() {
		return None
	}

	candle := pd.GetCandle(index)

	if p := detectSingleCandlePattern(candle); p != None {
		return p
	}

	if p := pd.detectTwoCandlePattern(index, candle); p != None {
		return p
	}

	if p := pd.detectThreeCandlePattern(index, candle); p != None {
		return p
	}

	return None
}

func detectSingleCandlePattern(c Candle) Pattern {
	if c.isDoji() {
		if c.isDragonflyDoji() {
			return DragonflyDoji
		}
		if c.isGravestoneDoji() {
			return GravestoneDoji
		}
		return Doji
	}
	if c.isHammer() {
		return Hammer
	}
	if c.isInvertedHammer() {
		return InvertedHammer
	}
	if c.isMarubozu() {
		return Marubozu
	}
	if c.isSpinningTop() {
		return SpinningTop
	}
	return None
}

func (pd *PatternDetector) detectTwoCandlePattern(index int, current Candle) Pattern {
	if index < 1 {
		return None
	}

	prev := pd.GetCandle(index - 1)
	if current.isBullishEngulfing(prev) {
		return BullishEngulfing
	}
	if current.isBearishEngulfing(prev) {
		return BearishEngulfing
	}
	if current.isBullishHarami(prev) {
		return BullishHarami
	}
	if current.isBearishHarami(prev) {
		return BearishHarami
	}
	if current.isPiercingLine(prev) {
		return PiercingLine
	}
	if current.isDarkCloudCover(prev) {
		return DarkCloudCover
	}
	return None
}

func (pd *PatternDetector) detectThreeCandlePattern(index int, current Candle) Pattern {
	if index < 2 {
		return None
	}

	first := pd.GetCandle(index - 2)
	middle := pd.GetCandle(index - 1)
	if current.isMorningStar(first, middle) {
		return MorningStar
	}
	if current.isEveningStar(first, middle) {
		return EveningStar
	}
	return None
}

func (c Candle) body() decimal.Decimal {
	return c.Close.Sub(c.Open).Abs()
}

func (c Candle) upperShadow() decimal.Decimal {
	return c.High.Sub(c.Close.Max(c.Open))
}

func (c Candle) lowerShadow() decimal.Decimal {
	return c.Open.Min(c.Close).Sub(c.Low)
}

func (c Candle) candleRange() decimal.Decimal {
	return c.High.Sub(c.Low)
}

func (c Candle) isDoji() bool {
	body := c.body()
	rangeVal := c.candleRange()
	if rangeVal.IsZero() {
		return false
	}
	return body.Div(rangeVal).LT(decimal.New(0.1))
}

func (c Candle) isDragonflyDoji() bool {
	if !c.isDoji() {
		return false
	}
	upper := c.upperShadow()
	lower := c.lowerShadow()
	return lower.GT(upper.Mul(decimal.New(2)))
}

func (c Candle) isGravestoneDoji() bool {
	if !c.isDoji() {
		return false
	}
	upper := c.upperShadow()
	lower := c.lowerShadow()
	return upper.GT(lower.Mul(decimal.New(2)))
}

func (c Candle) isHammer() bool {
	body := c.body()
	upper := c.upperShadow()
	lower := c.lowerShadow()
	rangeVal := c.candleRange()

	if rangeVal.IsZero() {
		return false
	}

	bodyRatio := body.Div(rangeVal)
	upperRatio := upper.Div(rangeVal)
	lowerRatio := lower.Div(rangeVal)
	bodyNearHigh := c.Open.Min(c.Close).GTE(c.Low.Add(rangeVal.Mul(decimal.New(0.5))))

	return lowerRatio.GTE(decimal.New(0.5)) &&
		upperRatio.LTE(decimal.New(0.2)) &&
		bodyRatio.LTE(decimal.New(0.35)) &&
		bodyNearHigh
}

func (c Candle) isInvertedHammer() bool {
	body := c.body()
	lower := c.lowerShadow()
	upper := c.upperShadow()
	rangeVal := c.candleRange()

	if rangeVal.IsZero() {
		return false
	}

	bodyRatio := body.Div(rangeVal)
	upperRatio := upper.Div(rangeVal)
	lowerRatio := lower.Div(rangeVal)
	bodyNearLow := c.Open.Max(c.Close).LTE(c.Low.Add(rangeVal.Mul(decimal.New(0.5))))

	return upperRatio.GTE(decimal.New(0.55)) &&
		lowerRatio.LTE(decimal.New(0.2)) &&
		bodyRatio.LTE(decimal.New(0.35)) &&
		bodyNearLow
}

func (c Candle) isSpinningTop() bool {
	body := c.body()
	rangeVal := c.candleRange()
	if rangeVal.IsZero() {
		return false
	}
	bodyRatio := body.Div(rangeVal)
	return bodyRatio.GT(decimal.New(0.1)) && bodyRatio.LT(decimal.New(0.3))
}

func (c Candle) isMarubozu() bool {
	body := c.body()
	rangeVal := c.candleRange()
	if rangeVal.IsZero() {
		return false
	}
	bodyRatio := body.Div(rangeVal)
	upper := c.upperShadow()
	lower := c.lowerShadow()
	return bodyRatio.GT(decimal.New(0.95)) && upper.LT(body.Mul(decimal.New(0.05))) && lower.LT(body.Mul(decimal.New(0.05)))
}

func (c Candle) isBullishEngulfing(prev Candle) bool {
	if prev.isDoji() || c.isDoji() {
		return false
	}

	currentBullish := c.Close.GT(c.Open)
	if !currentBullish {
		return false
	}

	prevMin := prev.Open.Min(prev.Close)
	prevMax := prev.Open.Max(prev.Close)
	return c.Open.LT(prevMin) && c.Close.GT(prevMax)
}

func (c Candle) isBearishEngulfing(prev Candle) bool {
	if prev.isDoji() || c.isDoji() {
		return false
	}

	currentBearish := c.Close.LT(c.Open)
	if !currentBearish {
		return false
	}

	prevMin := prev.Open.Min(prev.Close)
	prevMax := prev.Open.Max(prev.Close)
	return c.Open.GT(prevMax) && c.Close.LT(prevMin)
}

func (c Candle) isBullishHarami(prev Candle) bool {
	if prev.isDoji() || c.isDoji() {
		return false
	}

	prevBearish := prev.Close.LT(prev.Open)
	currentBullish := c.Close.GT(c.Open)

	if !prevBearish || !currentBullish {
		return false
	}

	return c.Open.GT(prev.Close) && c.Close.LT(prev.Open) && c.body().LT(prev.body().Mul(decimal.New(0.3)))
}

func (c Candle) isBearishHarami(prev Candle) bool {
	if prev.isDoji() || c.isDoji() {
		return false
	}

	prevBullish := prev.Close.GT(prev.Open)
	currentBearish := c.Close.LT(c.Open)

	if !prevBullish || !currentBearish {
		return false
	}

	return c.Open.LT(prev.Close) && c.Close.GT(prev.Open) && c.body().LT(prev.body().Mul(decimal.New(0.3)))
}

func (c Candle) isMorningStar(first, middle Candle) bool {
	if first.isDoji() || middle.isDoji() || c.isDoji() {
		return false
	}

	firstBearish := first.Close.LT(first.Open)
	middleSmall := middle.body().LT(first.body().Mul(decimal.New(0.3)))
	currentBullish := c.Close.GT(c.Open)

	middleInGap := middle.High.LT(first.Low) || middle.Low.GT(first.High)
	currentInGap := c.Open.GT(first.Close) && c.Close.GT(middle.High)

	return firstBearish && middleSmall && currentBullish && middleInGap && currentInGap
}

func (c Candle) isEveningStar(first, middle Candle) bool {
	if first.isDoji() || middle.isDoji() || c.isDoji() {
		return false
	}

	firstBullish := first.Close.GT(first.Open)
	middleSmall := middle.body().LT(first.body().Mul(decimal.New(0.3)))
	currentBearish := c.Close.LT(c.Open)

	middleInGap := middle.High.LT(first.Low) || middle.Low.GT(first.High)
	currentInGap := c.Open.LT(first.Close) && c.Close.LT(middle.Low)

	return firstBullish && middleSmall && currentBearish && middleInGap && currentInGap
}

func (c Candle) isPiercingLine(prev Candle) bool {
	if prev.isDoji() || c.isDoji() {
		return false
	}

	prevBearish := prev.Close.LT(prev.Open)
	currentBullish := c.Close.GT(c.Open)

	if !prevBearish || !currentBullish {
		return false
	}

	openBelow := c.Open.LT(prev.Low)
	midpoint := prev.Open.Add(prev.Close).Div(decimal.New(2))
	closesAboveMidpoint := c.Close.GT(midpoint)
	closesBelowOpen := c.Close.LT(prev.Open)

	return openBelow && closesAboveMidpoint && closesBelowOpen
}

func (c Candle) isDarkCloudCover(prev Candle) bool {
	if prev.isDoji() || c.isDoji() {
		return false
	}

	prevBullish := prev.Close.GT(prev.Open)
	currentBearish := c.Close.LT(c.Open)

	if !prevBullish || !currentBearish {
		return false
	}

	openAbove := c.Open.GT(prev.High)
	midpoint := prev.Open.Add(prev.Close).Div(decimal.New(2))
	closesBelowMidpoint := c.Close.LT(midpoint)
	closesBelowPrevOpen := c.Close.LT(prev.Open)

	return openAbove && closesBelowMidpoint && closesBelowPrevOpen
}

var patternNames = map[Pattern]string{
	Doji:               "Doji",
	DragonflyDoji:      "Dragonfly Doji",
	GravestoneDoji:     "Gravestone Doji",
	Hammer:             "Hammer",
	InvertedHammer:     "Inverted Hammer",
	HangingMan:         "Hanging Man",
	ShootingStar:       "Shooting Star",
	BullishEngulfing:   "Bullish Engulfing",
	BearishEngulfing:   "Bearish Engulfing",
	BullishHarami:      "Bullish Harami",
	BearishHarami:      "Bearish Harami",
	MorningStar:        "Morning Star",
	EveningStar:        "Evening Star",
	ThreeWhiteSoldiers: "Three White Soldiers",
	ThreeBlackCrows:    "Three Black Crows",
	SpinningTop:        "Spinning Top",
	Marubozu:           "Marubozu",
	DojiStar:           "Doji Star",
	PiercingLine:       "Piercing Line",
	DarkCloudCover:     "Dark Cloud Cover",
}

func (p Pattern) String() string {
	if s, ok := patternNames[p]; ok {
		return s
	}
	return "None"
}
