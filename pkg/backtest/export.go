package backtest

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/irfndi/goflux/pkg/metrics"
)

type ExportFormat string

const (
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatJSON ExportFormat = "json"
)

type ExportOptions struct {
	Format        ExportFormat
	TimeFormat    string
	PrettyPrint   bool
	IncludeHeader bool
}

func DefaultExportOptions() ExportOptions {
	return ExportOptions{
		Format:        ExportFormatCSV,
		TimeFormat:    time.RFC3339,
		PrettyPrint:   false,
		IncludeHeader: true,
	}
}

type BacktestExporter struct {
	opts ExportOptions
}

func NewBacktestExporter(opts ExportOptions) *BacktestExporter {
	return &BacktestExporter{opts: opts}
}

func (e *BacktestExporter) Export(result BacktestResult, writer io.Writer) error {
	switch e.opts.Format {
	case ExportFormatCSV:
		return e.exportCSV(result, writer)
	case ExportFormatJSON:
		return e.exportJSON(result, writer)
	default:
		return fmt.Errorf("unsupported export format: %s", e.opts.Format)
	}
}

func (e *BacktestExporter) ExportTrades(trades []Trade, writer io.Writer) error {
	switch e.opts.Format {
	case ExportFormatCSV:
		return e.exportTradesCSV(trades, writer)
	case ExportFormatJSON:
		return e.exportTradesJSON(trades, writer)
	default:
		return fmt.Errorf("unsupported export format: %s", e.opts.Format)
	}
}

func (e *BacktestExporter) ExportEquityCurve(curve []metrics.EquityPoint, writer io.Writer) error {
	switch e.opts.Format {
	case ExportFormatCSV:
		return e.exportEquityCurveCSV(curve, writer)
	case ExportFormatJSON:
		return e.exportEquityCurveJSON(curve, writer)
	default:
		return fmt.Errorf("unsupported export format: %s", e.opts.Format)
	}
}

func (e *BacktestExporter) ExportSummary(result BacktestResult, writer io.Writer) error {
	switch e.opts.Format {
	case ExportFormatCSV:
		return e.exportSummaryCSV(result, writer)
	case ExportFormatJSON:
		return e.exportSummaryJSON(result, writer)
	default:
		return fmt.Errorf("unsupported export format: %s", e.opts.Format)
	}
}

func (e *BacktestExporter) ExportToFile(result BacktestResult, path string) (err error) {
	file, err := os.Create(filepath.Clean(path))
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	return e.Export(result, file)
}

// --- CSV exporters ---

func (e *BacktestExporter) exportCSV(result BacktestResult, writer io.Writer) error {
	if err := e.exportSummaryCSV(result, writer); err != nil {
		return err
	}
	if _, err := io.WriteString(writer, "\n"); err != nil {
		return err
	}
	if err := e.exportTradesCSV(result.Trades, writer); err != nil {
		return err
	}
	return nil
}

func (e *BacktestExporter) exportTradesCSV(trades []Trade, writer io.Writer) error {
	w := csv.NewWriter(writer)
	defer w.Flush()

	if e.opts.IncludeHeader {
		if err := w.Write([]string{"entry_time", "exit_time", "entry_price", "exit_price", "direction", "quantity", "profit", "profit_percent", "duration"}); err != nil {
			return err
		}
	}

	for _, t := range trades {
		record := []string{
			strconv.Itoa(t.EntryTime),
			strconv.Itoa(t.ExitTime),
			t.EntryPrice.String(),
			t.ExitPrice.String(),
			t.Direction,
			t.Quantity.String(),
			t.Profit.String(),
			t.ProfitPercent.String(),
			strconv.Itoa(t.Duration),
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}
	return nil
}

func (e *BacktestExporter) exportEquityCurveCSV(curve []metrics.EquityPoint, writer io.Writer) error {
	w := csv.NewWriter(writer)
	defer w.Flush()

	if e.opts.IncludeHeader {
		if err := w.Write([]string{"equity", "drawdown", "drawdown_percent"}); err != nil {
			return err
		}
	}

	for _, point := range curve {
		record := []string{
			point.Equity.String(),
			point.Drawdown.String(),
			point.DrawdownPct.String(),
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}
	return nil
}

func (e *BacktestExporter) exportSummaryCSV(result BacktestResult, writer io.Writer) error {
	w := csv.NewWriter(writer)
	defer w.Flush()

	if e.opts.IncludeHeader {
		if err := w.Write([]string{"metric", "value"}); err != nil {
			return err
		}
	}

	rows := []struct{ metric, value string }{
		{"total_trades", strconv.Itoa(result.TotalTrades)},
		{"winning_trades", strconv.Itoa(result.WinningTrades)},
		{"losing_trades", strconv.Itoa(result.LosingTrades)},
		{"win_rate", result.WinRate.String()},
		{"total_profit", result.TotalProfit.String()},
		{"total_loss", result.TotalLoss.String()},
		{"net_profit", result.NetProfit.String()},
		{"gross_profit", result.GrossProfit.String()},
		{"gross_loss", result.GrossLoss.String()},
		{"profit_factor", result.ProfitFactor.String()},
		{"average_win", result.AverageWin.String()},
		{"average_loss", result.AverageLoss.String()},
		{"average_trade", result.AverageTrade.String()},
		{"max_consecutive_wins", strconv.Itoa(result.MaxConsecutiveWins)},
		{"max_consecutive_losses", strconv.Itoa(result.MaxConsecutiveLosses)},
		{"max_drawdown", result.MaxDrawdown.String()},
		{"max_drawdown_percent", result.MaxDrawdownPercent.String()},
		{"recovery_factor", result.RecoveryFactor.String()},
		{"risk_reward_ratio", result.RiskRewardRatio.String()},
		{"calmar_ratio", result.CalmarRatio.String()},
		{"sortino_ratio", result.SortinoRatio.String()},
		{"sharpe_ratio", result.SharpeRatio.String()},
		{"cagr", result.CAGR.String()},
		{"final_equity", result.FinalEquity.String()},
		{"initial_capital", result.InitialCapital.String()},
	}

	for _, row := range rows {
		if err := w.Write([]string{row.metric, row.value}); err != nil {
			return err
		}
	}
	return nil
}

// --- JSON exporters ---

func (e *BacktestExporter) exportJSON(result BacktestResult, writer io.Writer) error {
	summary := e.buildSummaryMap(result)
	output := map[string]interface{}{
		"summary":  summary,
		"trades":   result.Trades,
		"analysis": result.Analysis,
	}

	enc := json.NewEncoder(writer)
	if e.opts.PrettyPrint {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(output)
}

func (e *BacktestExporter) exportTradesJSON(trades []Trade, writer io.Writer) error {
	enc := json.NewEncoder(writer)
	if e.opts.PrettyPrint {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(trades)
}

func (e *BacktestExporter) exportEquityCurveJSON(curve []metrics.EquityPoint, writer io.Writer) error {
	enc := json.NewEncoder(writer)
	if e.opts.PrettyPrint {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(curve)
}

func (e *BacktestExporter) exportSummaryJSON(result BacktestResult, writer io.Writer) error {
	summary := e.buildSummaryMap(result)
	enc := json.NewEncoder(writer)
	if e.opts.PrettyPrint {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(summary)
}

func (e *BacktestExporter) buildSummaryMap(result BacktestResult) map[string]interface{} {
	return map[string]interface{}{
		"total_trades":           result.TotalTrades,
		"winning_trades":         result.WinningTrades,
		"losing_trades":          result.LosingTrades,
		"win_rate":               result.WinRate,
		"total_profit":           result.TotalProfit,
		"total_loss":             result.TotalLoss,
		"net_profit":             result.NetProfit,
		"gross_profit":           result.GrossProfit,
		"gross_loss":             result.GrossLoss,
		"profit_factor":          result.ProfitFactor,
		"average_win":            result.AverageWin,
		"average_loss":           result.AverageLoss,
		"average_trade":          result.AverageTrade,
		"max_consecutive_wins":   result.MaxConsecutiveWins,
		"max_consecutive_losses": result.MaxConsecutiveLosses,
		"max_drawdown":           result.MaxDrawdown,
		"max_drawdown_percent":   result.MaxDrawdownPercent,
		"recovery_factor":        result.RecoveryFactor,
		"risk_reward_ratio":      result.RiskRewardRatio,
		"calmar_ratio":           result.CalmarRatio,
		"sortino_ratio":          result.SortinoRatio,
		"sharpe_ratio":           result.SharpeRatio,
		"cagr":                   result.CAGR,
		"final_equity":           result.FinalEquity,
		"initial_capital":        result.InitialCapital,
	}
}
