package database

import "fmt"

// InfluxDBStorage implements Storage using the historical InfluxDB API.
//
// Deprecated: the backend remains a compatibility scaffold and is not
// connected to an InfluxDB client. Use Storage with an application-owned
// implementation for production persistence.
type InfluxDBStorage struct {
	dsn       string
	database  string
	precision string
}

// NewInfluxDBStorage creates a compatibility InfluxDB storage instance.
func NewInfluxDBStorage(dsn, database string) (*InfluxDBStorage, error) {
	return &InfluxDBStorage{dsn: dsn, database: database, precision: "ns"}, nil
}

// StoreCandle stores a single candle in InfluxDB.
func (s *InfluxDBStorage) StoreCandle(symbol string, candle *Candle) error {
	return fmt.Errorf("InfluxDBStorage.StoreCandle not yet implemented")
}

// StoreCandles stores multiple candles in InfluxDB.
func (s *InfluxDBStorage) StoreCandles(symbol string, candles []*Candle) error {
	return fmt.Errorf("InfluxDBStorage.StoreCandles not yet implemented")
}

// GetCandles retrieves candles from InfluxDB.
func (s *InfluxDBStorage) GetCandles(symbol string, startTime, endTime int64) ([]*Candle, error) {
	return nil, fmt.Errorf("InfluxDBStorage.GetCandles not yet implemented")
}

// GetLatestCandles retrieves recent candles from InfluxDB.
func (s *InfluxDBStorage) GetLatestCandles(symbol string, limit int) ([]*Candle, error) {
	return nil, fmt.Errorf("InfluxDBStorage.GetLatestCandles not yet implemented")
}

// DeleteSymbol removes all data for a symbol from InfluxDB.
func (s *InfluxDBStorage) DeleteSymbol(symbol string) error {
	return fmt.Errorf("InfluxDBStorage.DeleteSymbol not yet implemented")
}

// Close closes the InfluxDB connection.
func (s *InfluxDBStorage) Close() error {
	return fmt.Errorf("InfluxDBStorage.Close not yet implemented")
}
