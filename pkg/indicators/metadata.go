package indicators

import (
	"fmt"
	"sync"
)

// IndicatorMetadata describes an indicator's properties
type IndicatorMetadata struct {
	Name        string
	Category    string
	Description string
	Inputs      []string
	Lookback    int
}

var metadataRegistry = make(map[string]IndicatorMetadata)
var metadataMu sync.RWMutex

// RegisterMetadata registers metadata for an indicator name
func RegisterMetadata(name string, meta IndicatorMetadata) {
	meta.Inputs = append([]string(nil), meta.Inputs...)
	metadataMu.Lock()
	defer metadataMu.Unlock()
	metadataRegistry[name] = meta
}

// GetMetadata returns metadata for an indicator name
func GetMetadata(name string) (IndicatorMetadata, error) {
	metadataMu.RLock()
	meta, ok := metadataRegistry[name]
	metadataMu.RUnlock()
	if !ok {
		return IndicatorMetadata{}, fmt.Errorf("metadata for %s not found", name)
	}
	meta.Inputs = append([]string(nil), meta.Inputs...)
	return meta, nil
}

func init() {
	RegisterMetadata("sma", IndicatorMetadata{
		Name:     "Simple Moving Average",
		Category: "Overlap Studies",
		Lookback: 0, // Varies by period
	})
	RegisterMetadata("ema", IndicatorMetadata{
		Name:     "Exponential Moving Average",
		Category: "Overlap Studies",
	})
	RegisterMetadata("rsi", IndicatorMetadata{
		Name:     "Relative Strength Index",
		Category: "Momentum Indicators",
	})
}
