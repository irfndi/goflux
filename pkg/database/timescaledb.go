package database

import "fmt"

// TimescaleDBStorage implements Storage using the historical TimescaleDB API.
//
// Deprecated: the backend remains a compatibility scaffold and is not
// connected to a PostgreSQL client. Use Storage with an application-owned
// implementation for production persistence.
type TimescaleDBStorage struct {
	dsn string
}

// NewTimescaleDBStorage creates a compatibility TimescaleDB storage instance.
func NewTimescaleDBStorage(dsn string) (*TimescaleDBStorage, error) {
	return &TimescaleDBStorage{dsn: dsn}, nil
}

// StoreCandle stores a single candle in TimescaleDB.
func (s *TimescaleDBStorage) StoreCandle(symbol string, candle *Candle) error {
	return fmt.Errorf("TimescaleDBStorage.StoreCandle not yet implemented")
}

// StoreCandles stores multiple candles in TimescaleDB.
func (s *TimescaleDBStorage) StoreCandles(symbol string, candles []*Candle) error {
	return fmt.Errorf("TimescaleDBStorage.StoreCandles not yet implemented")
}

// GetCandles retrieves candles from TimescaleDB.
func (s *TimescaleDBStorage) GetCandles(symbol string, startTime, endTime int64) ([]*Candle, error) {
	return nil, fmt.Errorf("TimescaleDBStorage.GetCandles not yet implemented")
}

// GetLatestCandles retrieves recent candles from TimescaleDB.
func (s *TimescaleDBStorage) GetLatestCandles(symbol string, limit int) ([]*Candle, error) {
	return nil, fmt.Errorf("TimescaleDBStorage.GetLatestCandles not yet implemented")
}

// DeleteSymbol removes all data for a symbol from TimescaleDB.
func (s *TimescaleDBStorage) DeleteSymbol(symbol string) error {
	return fmt.Errorf("TimescaleDBStorage.DeleteSymbol not yet implemented")
}

// Close closes the TimescaleDB connection.
func (s *TimescaleDBStorage) Close() error {
	return fmt.Errorf("TimescaleDBStorage.Close not yet implemented")
}
