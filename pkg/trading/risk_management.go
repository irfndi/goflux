package trading

import (
	"github.com/irfndi/goflux/pkg/decimal"
)

type DrawdownProtectionRule struct {
	MaxDrawdownPct decimal.Decimal
	InitialCapital decimal.Decimal
}

func NewDrawdownProtectionRule(maxDrawdownPct float64) Rule {
	return &DrawdownProtectionRule{
		MaxDrawdownPct: decimal.New(maxDrawdownPct),
		InitialCapital: decimal.New(10000),
	}
}

func NewDrawdownProtectionRuleWithCapital(maxDrawdownPct float64, initialCapital float64) Rule {
	return &DrawdownProtectionRule{
		MaxDrawdownPct: decimal.New(maxDrawdownPct),
		InitialCapital: decimal.New(initialCapital),
	}
}

func (dpr *DrawdownProtectionRule) IsSatisfied(index int, record *TradingRecord) bool {
	if len(record.Trades) == 0 && record.CurrentPosition().IsNew() {
		return false
	}

	peakBalance := dpr.InitialCapital
	currentBalance := dpr.InitialCapital

	for _, trade := range record.Trades {
		if trade.IsClosed() {
			pnl := trade.ExitValue().Sub(trade.CostBasis())
			currentBalance = currentBalance.Add(pnl)
			if currentBalance.GT(peakBalance) {
				peakBalance = currentBalance
			}
		}
	}

	drawdown := peakBalance.Sub(currentBalance)
	if peakBalance.IsZero() {
		return false
	}

	drawdownPct := drawdown.Div(peakBalance)
	return drawdownPct.GTE(dpr.MaxDrawdownPct)
}

type MaxLossRule struct {
	MaxLossPct     decimal.Decimal
	InitialCapital decimal.Decimal
}

func NewMaxLossRule(maxLossPct float64) Rule {
	return MaxLossRule{
		MaxLossPct:     decimal.New(maxLossPct),
		InitialCapital: decimal.New(10000),
	}
}

func NewMaxLossRuleWithCapital(maxLossPct float64, initialCapital float64) Rule {
	return MaxLossRule{
		MaxLossPct:     decimal.New(maxLossPct),
		InitialCapital: decimal.New(initialCapital),
	}
}

func (mlr MaxLossRule) IsSatisfied(index int, record *TradingRecord) bool {
	if record.CurrentPosition().IsNew() && len(record.Trades) == 0 {
		return false
	}

	var totalPL decimal.Decimal
	for _, trade := range record.Trades {
		if trade.IsClosed() {
			pnl := trade.ExitValue().Sub(trade.CostBasis())
			totalPL = totalPL.Add(pnl)
		}
	}

	maxAllowedLoss := mlr.InitialCapital.Mul(mlr.MaxLossPct)
	currentLoss := mlr.InitialCapital.Sub(mlr.InitialCapital.Add(totalPL))

	return currentLoss.GTE(maxAllowedLoss)
}

type DailyLossLimitRule struct {
	MaxDailyLoss decimal.Decimal
	DailyPnL     decimal.Decimal
	SessionStart decimal.Decimal
}

func NewDailyLossLimitRule(maxDailyLoss float64) Rule {
	return DailyLossLimitRule{
		MaxDailyLoss: decimal.New(maxDailyLoss),
		DailyPnL:     decimal.ZERO,
		SessionStart: decimal.New(10000),
	}
}

func (dllr DailyLossLimitRule) IsSatisfied(index int, record *TradingRecord) bool {
	if len(record.Trades) == 0 {
		return false
	}

	var sessionPL decimal.Decimal
	for _, trade := range record.Trades {
		if trade.IsClosed() {
			pnl := trade.ExitValue().Sub(trade.CostBasis())
			sessionPL = sessionPL.Add(pnl)
		}
	}

	return sessionPL.LTE(dllr.MaxDailyLoss.Neg())
}

type ConsecutiveLossRule struct {
	MaxConsecutiveLosses int
}

func NewConsecutiveLossRule(maxConsecutiveLosses int) Rule {
	return &ConsecutiveLossRule{
		MaxConsecutiveLosses: maxConsecutiveLosses,
	}
}

func (clr *ConsecutiveLossRule) IsSatisfied(index int, record *TradingRecord) bool {
	if len(record.Trades) == 0 {
		return false
	}

	consecutive := 0
	for i := len(record.Trades) - 1; i >= 0; i-- {
		trade := record.Trades[i]
		if !trade.IsClosed() {
			continue
		}

		var pnl decimal.Decimal
		if trade.IsLong() {
			pnl = trade.ExitValue().Sub(trade.CostBasis())
		} else {
			pnl = trade.CostBasis().Sub(trade.ExitValue())
		}

		if pnl.LT(decimal.ZERO) {
			consecutive++
			if consecutive >= clr.MaxConsecutiveLosses {
				return true
			}
		} else {
			break
		}
	}

	return false
}

type PositionSizeRiskRule struct {
	MaxPositionSize decimal.Decimal
	CurrentSize     decimal.Decimal
}

func NewPositionSizeRiskRule(maxPositionSize float64) Rule {
	return PositionSizeRiskRule{
		MaxPositionSize: decimal.New(maxPositionSize),
		CurrentSize:     decimal.ZERO,
	}
}

func (psrr PositionSizeRiskRule) IsSatisfied(index int, record *TradingRecord) bool {
	if record.CurrentPosition().IsNew() {
		return false
	}

	psrr.CurrentSize = record.CurrentPosition().CostBasis()
	return psrr.CurrentSize.GT(psrr.MaxPositionSize)
}

type PortfolioExposureRule struct {
	MaxExposure decimal.Decimal
	TotalValue  decimal.Decimal
}

func NewPortfolioExposureRule(maxExposure float64) Rule {
	return PortfolioExposureRule{
		MaxExposure: decimal.New(maxExposure),
		TotalValue:  decimal.New(10000),
	}
}

func (per PortfolioExposureRule) IsSatisfied(index int, record *TradingRecord) bool {
	var totalExposure decimal.Decimal
	if record.CurrentPosition().IsOpen() {
		totalExposure = totalExposure.Add(record.CurrentPosition().CostBasis())
	}
	for _, trade := range record.Trades {
		if trade.IsOpen() {
			totalExposure = totalExposure.Add(trade.CostBasis())
		}
	}

	exposurePct := totalExposure.Div(per.TotalValue)
	return exposurePct.GTE(per.MaxExposure)
}

func (mlr *MaxLossRule) Reset() {
	mlr.InitialCapital = decimal.New(10000)
}
