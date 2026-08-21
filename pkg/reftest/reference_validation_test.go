package reftest

import "testing"

func TestValidateAgainstReference(t *testing.T) {
	tests := []struct {
		name       string
		calculated float64
		expected   float64
		tolerance  float64
		want       bool
	}{
		{name: "exact match", calculated: 100, expected: 100, tolerance: 0.01, want: true},
		{name: "within tolerance", calculated: 100.5, expected: 100, tolerance: 0.01, want: true},
		{name: "outside tolerance", calculated: 102, expected: 100, tolerance: 0.01, want: false},
		{name: "default tolerance", calculated: 101, expected: 100, tolerance: 0, want: true},
		{name: "negative tolerance uses magnitude", calculated: 100.5, expected: 100, tolerance: -0.01, want: true},
		{name: "zero reference within tolerance", calculated: 0.005, expected: 0, tolerance: 0.01, want: true},
		{name: "zero reference outside tolerance", calculated: 0.02, expected: 0, tolerance: 0.01, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateAgainstReference(tt.calculated, tt.expected, tt.tolerance); got != tt.want {
				t.Fatalf("ValidateAgainstReference(%v, %v, %v) = %v, want %v", tt.calculated, tt.expected, tt.tolerance, got, tt.want)
			}
		})
	}
}
