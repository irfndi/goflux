package database

import "testing"

var (
	_ Storage = (*InfluxDBStorage)(nil)
	_ Storage = (*TimescaleDBStorage)(nil)
)

func TestLegacyStorageConstructorsRemainAvailable(t *testing.T) {
	influx, err := NewInfluxDBStorage("http://localhost:8086", "trading")
	if err != nil || influx == nil {
		t.Fatalf("NewInfluxDBStorage() = %v, %v", influx, err)
	}
	timescale, err := NewTimescaleDBStorage("postgres://localhost/trading")
	if err != nil || timescale == nil {
		t.Fatalf("NewTimescaleDBStorage() = %v, %v", timescale, err)
	}
}
