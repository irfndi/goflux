package series

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/irfndi/goflux/pkg/decimal"
)

// CSVConfig describes how to parse a CSV file into a TimeSeries
type CSVConfig struct {
	TimeFormat  string
	TimeIndex   int
	OpenIndex   int
	HighIndex   int
	LowIndex    int
	CloseIndex  int
	VolumeIndex int
	HasHeader   bool
}

// NewCSVConfig returns a default CSVConfig with standard indices
func NewCSVConfig() CSVConfig {
	return CSVConfig{
		TimeFormat:  time.RFC3339,
		TimeIndex:   0,
		OpenIndex:   1,
		HighIndex:   2,
		LowIndex:    3,
		CloseIndex:  4,
		VolumeIndex: 5,
		HasHeader:   true,
	}
}

// LoadCSV parses CSV data from an io.Reader and returns a TimeSeries
func LoadCSV(reader io.Reader, config CSVConfig) (*TimeSeries, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}

	csvReader := csv.NewReader(reader)
	if config.HasHeader {
		_, err := csvReader.Read() // Skip header
		if err != nil {
			return nil, fmt.Errorf("error reading CSV header: %w", err)
		}
	}

	ts := NewTimeSeries()
	rowNumber := 1
	if config.HasHeader {
		rowNumber++
	}
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading CSV record on row %d: %w", rowNumber, err)
		}
		if required := config.maxIndex(); required >= len(record) {
			return nil, fmt.Errorf("CSV row %d has %d fields; index %d is out of range", rowNumber, len(record), required)
		}

		t, err := time.Parse(config.TimeFormat, record[config.TimeIndex])
		if err != nil {
			return nil, fmt.Errorf("error parsing time on row %d: %w", rowNumber, err)
		}

		// Calculate duration if possible, otherwise assume 1 min or similar?
		// Better to let user specify or infer from first two candles.
		// For now, we'll use a fixed duration of 0 and let users use resample or fix it later.
		// Actually, NewTimePeriod needs a duration.

		candle := NewCandle(NewTimePeriod(t, 0)) // Initial duration 0
		var parseDecimal = func(label string, index int) (decimal.Decimal, error) {
			value, parseErr := decimal.NewFromStringWithError(record[index])
			if parseErr != nil {
				return decimal.ZERO, fmt.Errorf("error parsing %s on row %d: %w", label, rowNumber, parseErr)
			}
			return value, nil
		}
		if candle.OpenPrice, err = parseDecimal("open", config.OpenIndex); err != nil {
			return nil, err
		}
		if candle.MaxPrice, err = parseDecimal("high", config.HighIndex); err != nil {
			return nil, err
		}
		if candle.MinPrice, err = parseDecimal("low", config.LowIndex); err != nil {
			return nil, err
		}
		if candle.ClosePrice, err = parseDecimal("close", config.CloseIndex); err != nil {
			return nil, err
		}
		if config.VolumeIndex >= 0 && config.VolumeIndex < len(record) {
			if candle.Volume, err = parseDecimal("volume", config.VolumeIndex); err != nil {
				return nil, err
			}
		}

		if err := ts.AddCandleErr(candle); err != nil {
			return nil, fmt.Errorf("error adding CSV row %d: %w", rowNumber, err)
		}
		rowNumber++
	}

	// Post-process to fix durations if we have at least 2 candles
	if ts.Length() >= 2 {
		d := ts.GetCandle(1).Period.Start.Sub(ts.GetCandle(0).Period.Start)
		for i := 0; i < ts.Length(); i++ {
			candle := ts.GetCandle(i)
			candle.Period.End = candle.Period.Start.Add(d)
		}
	} else if ts.Length() == 1 {
		// Default to 1 minute if only one candle?
		candle := ts.GetCandle(0)
		candle.Period.End = candle.Period.Start.Add(time.Minute)
	}

	return ts, nil
}

// JSONCandle represents a single candle in JSON format
type JSONCandle struct {
	Time   string  `json:"time"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
}

// LoadJSON parses JSON data from an io.Reader and returns a TimeSeries
func LoadJSON(reader io.Reader, timeFormat string) (*TimeSeries, error) {
	var jsonCandles []JSONCandle
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&jsonCandles); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %w", err)
	}

	ts := NewTimeSeries()
	for _, jc := range jsonCandles {
		t, err := time.Parse(timeFormat, jc.Time)
		if err != nil {
			return nil, fmt.Errorf("error parsing time %s: %w", jc.Time, err)
		}

		candle := NewCandle(NewTimePeriod(t, 0))
		candle.OpenPrice = decimal.New(jc.Open)
		candle.MaxPrice = decimal.New(jc.High)
		candle.MinPrice = decimal.New(jc.Low)
		candle.ClosePrice = decimal.New(jc.Close)
		candle.Volume = decimal.New(jc.Volume)

		if err := ts.AddCandleErr(candle); err != nil {
			return nil, fmt.Errorf("error adding JSON candle %q: %w", jc.Time, err)
		}
	}

	// Post-process to fix durations
	if ts.Length() >= 2 {
		d := ts.GetCandle(1).Period.Start.Sub(ts.GetCandle(0).Period.Start)
		for i := 0; i < ts.Length(); i++ {
			candle := ts.GetCandle(i)
			candle.Period.End = candle.Period.Start.Add(d)
		}
	} else if ts.Length() == 1 {
		candle := ts.GetCandle(0)
		candle.Period.End = candle.Period.Start.Add(time.Minute)
	}

	return ts, nil
}

func (config CSVConfig) validate() error {
	if config.TimeFormat == "" {
		return fmt.Errorf("CSV time format cannot be empty")
	}
	for name, index := range map[string]int{
		"time": config.TimeIndex, "open": config.OpenIndex, "high": config.HighIndex,
		"low": config.LowIndex, "close": config.CloseIndex,
	} {
		if index < 0 {
			return fmt.Errorf("CSV %s index cannot be negative: %d", name, index)
		}
	}
	if config.VolumeIndex < -1 {
		return fmt.Errorf("CSV volume index cannot be less than -1: %d", config.VolumeIndex)
	}
	return nil
}

func (config CSVConfig) maxIndex() int {
	maxIndex := config.TimeIndex
	for _, index := range []int{config.OpenIndex, config.HighIndex, config.LowIndex, config.CloseIndex} {
		if index > maxIndex {
			maxIndex = index
		}
	}
	return maxIndex
}
