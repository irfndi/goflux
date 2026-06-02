package backtest

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/irfndi/goflux/pkg/decimal"
	"github.com/irfndi/goflux/pkg/metrics"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func sampleResult() BacktestResult {
	return BacktestResult{
		TotalTrades:          2,
		WinningTrades:        1,
		LosingTrades:         1,
		WinRate:              decimal.New(0.5),
		TotalProfit:          decimal.New(100),
		TotalLoss:            decimal.New(-50),
		NetProfit:            decimal.New(50),
		GrossProfit:          decimal.New(100),
		GrossLoss:            decimal.New(-50),
		ProfitFactor:         decimal.New(2),
		AverageWin:           decimal.New(100),
		AverageLoss:          decimal.New(-50),
		AverageTrade:         decimal.New(25),
		MaxConsecutiveWins:   1,
		MaxConsecutiveLosses: 1,
		MaxDrawdown:          decimal.New(20),
		MaxDrawdownPercent:   decimal.New(0.02),
		RecoveryFactor:       decimal.New(2.5),
		RiskRewardRatio:      decimal.New(2),
		CalmarRatio:          decimal.New(1.5),
		SortinoRatio:         decimal.New(1.2),
		SharpeRatio:          decimal.New(1.1),
		CAGR:                 decimal.New(0.15),
		FinalEquity:          decimal.New(1050),
		InitialCapital:       decimal.New(1000),
		Trades: []Trade{
			{
				EntryTime:     0,
				EntryPrice:    decimal.New(100),
				ExitTime:      5,
				ExitPrice:     decimal.New(110),
				Direction:     "long",
				Quantity:      decimal.New(1),
				Profit:        decimal.New(10),
				ProfitPercent: decimal.New(0.1),
				Duration:      5,
			},
			{
				EntryTime:     6,
				EntryPrice:    decimal.New(110),
				ExitTime:      10,
				ExitPrice:     decimal.New(105),
				Direction:     "short",
				Quantity:      decimal.New(1),
				Profit:        decimal.New(5),
				ProfitPercent: decimal.New(0.045),
				Duration:      4,
			},
		},
		Analysis: AnalysisResult{"key": "value"},
	}
}

func sampleEquityCurve() []metrics.EquityPoint {
	return []metrics.EquityPoint{
		{Equity: decimal.New(1000), Drawdown: decimal.New(0), DrawdownPct: decimal.New(0)},
		{Equity: decimal.New(1050), Drawdown: decimal.New(0), DrawdownPct: decimal.New(0)},
		{Equity: decimal.New(1020), Drawdown: decimal.New(30), DrawdownPct: decimal.New(0.0285)},
	}
}

func TestDefaultExportOptions(t *testing.T) {
	opts := DefaultExportOptions()
	assert.Equal(t, ExportFormatCSV, opts.Format)
	assert.Equal(t, "2006-01-02T15:04:05Z07:00", opts.TimeFormat)
	assert.False(t, opts.PrettyPrint)
	assert.True(t, opts.IncludeHeader)
}

func TestExportCSV(t *testing.T) {
	result := sampleResult()
	buf := &bytes.Buffer{}
	exporter := NewBacktestExporter(DefaultExportOptions())

	err := exporter.Export(result, buf)
	require.NoError(t, err)

	output := buf.String()
	// Should contain summary header and trades header
	assert.Contains(t, output, "metric,value")
	assert.Contains(t, output, "entry_time,exit_time")
	assert.Contains(t, output, "total_trades,2")
	assert.Contains(t, output, "long")
	assert.Contains(t, output, "short")
}

func TestExportJSON(t *testing.T) {
	result := sampleResult()
	buf := &bytes.Buffer{}
	opts := DefaultExportOptions()
	opts.Format = ExportFormatJSON
	exporter := NewBacktestExporter(opts)

	err := exporter.Export(result, buf)
	require.NoError(t, err)

	var output map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)

	summary, ok := output["summary"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, float64(2), summary["total_trades"])

	trades, ok := output["trades"].([]interface{})
	require.True(t, ok)
	assert.Len(t, trades, 2)
}

func TestExportUnsupportedFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	opts := DefaultExportOptions()
	opts.Format = "xml"
	exporter := NewBacktestExporter(opts)

	err := exporter.Export(BacktestResult{}, buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported export format")
}

func TestExportTradesCSV(t *testing.T) {
	trades := sampleResult().Trades
	buf := &bytes.Buffer{}
	exporter := NewBacktestExporter(DefaultExportOptions())

	err := exporter.ExportTrades(trades, buf)
	require.NoError(t, err)

	reader := csv.NewReader(buf)
	records, err := reader.ReadAll()
	require.NoError(t, err)
	assert.Equal(t, 3, len(records)) // header + 2 trades
	assert.Equal(t, "entry_time", records[0][0])
	assert.Equal(t, "long", records[1][4])
	assert.Equal(t, "short", records[2][4])
}

func TestExportTradesCSVNoHeader(t *testing.T) {
	trades := sampleResult().Trades
	buf := &bytes.Buffer{}
	opts := DefaultExportOptions()
	opts.IncludeHeader = false
	exporter := NewBacktestExporter(opts)

	err := exporter.ExportTrades(trades, buf)
	require.NoError(t, err)

	reader := csv.NewReader(buf)
	records, err := reader.ReadAll()
	require.NoError(t, err)
	assert.Equal(t, 2, len(records)) // no header
}

func TestExportTradesJSON(t *testing.T) {
	trades := sampleResult().Trades
	buf := &bytes.Buffer{}
	opts := DefaultExportOptions()
	opts.Format = ExportFormatJSON
	exporter := NewBacktestExporter(opts)

	err := exporter.ExportTrades(trades, buf)
	require.NoError(t, err)

	var output []map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)
	assert.Len(t, output, 2)
	assert.Equal(t, float64(0), output[0]["EntryTime"])
}

func TestExportEquityCurveCSV(t *testing.T) {
	curve := sampleEquityCurve()
	buf := &bytes.Buffer{}
	exporter := NewBacktestExporter(DefaultExportOptions())

	err := exporter.ExportEquityCurve(curve, buf)
	require.NoError(t, err)

	reader := csv.NewReader(buf)
	records, err := reader.ReadAll()
	require.NoError(t, err)
	assert.Equal(t, 4, len(records)) // header + 3 points
	assert.Equal(t, "equity", records[0][0])
	assert.Equal(t, "1000", records[1][0])
}

func TestExportEquityCurveJSON(t *testing.T) {
	curve := sampleEquityCurve()
	buf := &bytes.Buffer{}
	opts := DefaultExportOptions()
	opts.Format = ExportFormatJSON
	exporter := NewBacktestExporter(opts)

	err := exporter.ExportEquityCurve(curve, buf)
	require.NoError(t, err)

	var output []map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)
	assert.Len(t, output, 3)
}

func TestExportSummaryCSV(t *testing.T) {
	result := sampleResult()
	buf := &bytes.Buffer{}
	exporter := NewBacktestExporter(DefaultExportOptions())

	err := exporter.ExportSummary(result, buf)
	require.NoError(t, err)

	reader := csv.NewReader(buf)
	records, err := reader.ReadAll()
	require.NoError(t, err)
	assert.Equal(t, 26, len(records)) // header + 25 metrics
	assert.Equal(t, "metric", records[0][0])
	assert.Equal(t, "total_trades", records[1][0])
	assert.Equal(t, "2", records[1][1])
}

func TestExportSummaryJSON(t *testing.T) {
	result := sampleResult()
	buf := &bytes.Buffer{}
	opts := DefaultExportOptions()
	opts.Format = ExportFormatJSON
	exporter := NewBacktestExporter(opts)

	err := exporter.ExportSummary(result, buf)
	require.NoError(t, err)

	var output map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &output)
	require.NoError(t, err)
	assert.Equal(t, float64(2), output["total_trades"])
	assert.Equal(t, "0.5", output["win_rate"])
}

func TestExportSummaryJSONPrettyPrint(t *testing.T) {
	result := sampleResult()
	buf := &bytes.Buffer{}
	opts := DefaultExportOptions()
	opts.Format = ExportFormatJSON
	opts.PrettyPrint = true
	exporter := NewBacktestExporter(opts)

	err := exporter.ExportSummary(result, buf)
	require.NoError(t, err)

	// Pretty-printed JSON contains indentation
	assert.True(t, strings.Contains(buf.String(), "\n"))
	assert.True(t, strings.Contains(buf.String(), "  "))
}

func TestExportToFile(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "result.csv")
	result := sampleResult()
	exporter := NewBacktestExporter(DefaultExportOptions())

	err := exporter.ExportToFile(result, path)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "total_trades,2")
}

func TestExportToFileUnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "result.xml")
	opts := DefaultExportOptions()
	opts.Format = "xml"
	exporter := NewBacktestExporter(opts)

	err := exporter.ExportToFile(BacktestResult{}, path)
	assert.Error(t, err)
}

func TestExportTradesEmpty(t *testing.T) {
	buf := &bytes.Buffer{}
	exporter := NewBacktestExporter(DefaultExportOptions())

	err := exporter.ExportTrades([]Trade{}, buf)
	require.NoError(t, err)

	reader := csv.NewReader(buf)
	records, err := reader.ReadAll()
	require.NoError(t, err)
	assert.Equal(t, 1, len(records)) // header only
}

func TestExportEquityCurveEmpty(t *testing.T) {
	buf := &bytes.Buffer{}
	exporter := NewBacktestExporter(DefaultExportOptions())

	err := exporter.ExportEquityCurve([]metrics.EquityPoint{}, buf)
	require.NoError(t, err)

	reader := csv.NewReader(buf)
	records, err := reader.ReadAll()
	require.NoError(t, err)
	assert.Equal(t, 1, len(records)) // header only
}

func TestExportSummaryEmptyResult(t *testing.T) {
	buf := &bytes.Buffer{}
	exporter := NewBacktestExporter(DefaultExportOptions())

	err := exporter.ExportSummary(BacktestResult{}, buf)
	require.NoError(t, err)

	reader := csv.NewReader(buf)
	records, err := reader.ReadAll()
	require.NoError(t, err)
	assert.Equal(t, 26, len(records))   // header + 25 metrics with zero values
	assert.Equal(t, "0", records[1][1]) // total_trades = 0
}
