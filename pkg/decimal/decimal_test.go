package decimal

import (
	"math"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected string
	}{
		{"zero", 0, "0"},
		{"positive", 123.456, "123.456"},
		{"negative", -123.456, "-123.456"},
		{"large", 1e10, "10000000000"},
		{"small", 1e-10, "0.0000000001"},
		{"pi", math.Pi, "3.141592653589793"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := New(tt.input)
			if d.String() != tt.expected {
				t.Errorf("New(%v).String() = %s, want %s", tt.input, d.String(), tt.expected)
			}
		})
	}
}

func TestNewFromInt(t *testing.T) {
	tests := []struct {
		input    int64
		expected float64
	}{
		{0, 0},
		{1, 1},
		{-1, -1},
		{123456789, 123456789},
		{-9876543210, -9876543210},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			d := NewFromInt(tt.input)
			if d.Float() != tt.expected {
				t.Errorf("NewFromInt(%d).Float() = %v, want %v", tt.input, d.Float(), tt.expected)
			}
		})
	}
}

func TestNewFromString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  float64
		wantPanic bool
	}{
		{"zero", "0", 0, false},
		{"positive", "123.456", 123.456, false},
		{"negative", "-123.456", -123.456, false},
		{"scientific", "1.23e10", 1.23e10, false},
		{"plus", "+123", 123, false},
		{"invalid", "abc", 0, true},
		{"empty", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Error("expected panic, got none")
					}
				}()
				NewFromString(tt.input)
			} else {
				d := NewFromString(tt.input)
				if d.Float() != tt.expected {
					t.Errorf("NewFromString(%q).Float() = %v, want %v", tt.input, d.Float(), tt.expected)
				}
			}
		})
	}
}

func TestNewFromStringWithError(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"valid integer", "123", "123", false},
		{"valid decimal", "123.456", "123.456", false},
		{"valid negative", "-123.456", "-123.456", false},
		{"valid scientific", "1.23e2", "123", false},
		{"invalid empty", "", "", true},
		{"invalid text", "abc", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := NewFromStringWithError(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewFromStringWithError(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && d.String() != tt.expected {
				t.Errorf("NewFromStringWithError(%q) = %s, want %s", tt.input, d.String(), tt.expected)
			}
		})
	}
}

func TestDecimalArithmetic(t *testing.T) {
	t.Run("Add", func(t *testing.T) {
		d1 := New(123.456)
		d2 := New(78.9)
		result := d1.Add(d2)
		expected := New(202.356)
		if result.Cmp(expected) != 0 {
			t.Errorf("Add() = %v, want %v", result, expected)
		}
	})

	t.Run("Sub", func(t *testing.T) {
		d1 := New(123.456)
		d2 := New(78.9)
		result := d1.Sub(d2)
		expected := New(44.556)
		if result.Cmp(expected) != 0 {
			t.Errorf("Sub() = %v, want %v", result, expected)
		}
	})

	t.Run("Mul", func(t *testing.T) {
		d1 := New(12.34)
		d2 := New(56.78)
		result := d1.Mul(d2)
		expected := New(12.34 * 56.78)
		if math.Abs(result.Float()-expected.Float()) > 1e-10 {
			t.Errorf("Mul() = %v, want %v", result, expected)
		}
	})

	t.Run("Div", func(t *testing.T) {
		d1 := New(100)
		d2 := New(8)
		result := d1.Div(d2)
		expected := New(12.5)
		if math.Abs(result.Float()-expected.Float()) > 1e-10 {
			t.Errorf("Div() = %v, want %v", result, expected)
		}
	})

	t.Run("DivByZero", func(t *testing.T) {
		d1 := New(100)
		d2 := ZERO
		result := d1.Div(d2)
		if !result.IsZero() {
			t.Errorf("DivByZero() = %v, want 0", result)
		}
	})
}

func TestDecimalComparison(t *testing.T) {
	d1 := New(100)
	d2 := New(100)
	d3 := New(200)

	t.Run("GT", func(t *testing.T) {
		if !d3.GT(d1) {
			t.Error("200 should be greater than 100")
		}
		if d1.GT(d2) {
			t.Error("100 should not be greater than 100")
		}
	})

	t.Run("GTE", func(t *testing.T) {
		if !d3.GTE(d1) {
			t.Error("200 should be greater than or equal to 100")
		}
		if !d1.GTE(d2) {
			t.Error("100 should be greater than or equal to 100")
		}
		if d1.GTE(d3) {
			t.Error("100 should not be greater than or equal to 200")
		}
	})

	t.Run("LT", func(t *testing.T) {
		if !d1.LT(d3) {
			t.Error("100 should be less than 200")
		}
		if d1.LT(d2) {
			t.Error("100 should not be less than 100")
		}
	})

	t.Run("LTE", func(t *testing.T) {
		if !d1.LTE(d3) {
			t.Error("100 should be less than or equal to 200")
		}
		if !d1.LTE(d2) {
			t.Error("100 should be less than or equal to 100")
		}
		if d3.LTE(d1) {
			t.Error("200 should not be less than or equal to 100")
		}
	})

	t.Run("EQ", func(t *testing.T) {
		if !d1.EQ(d2) {
			t.Error("100 should equal 100")
		}
		if d1.EQ(d3) {
			t.Error("100 should not equal 200")
		}
	})
}

func TestDecimalCmp(t *testing.T) {
	d1 := New(100)
	d2 := New(100)
	d3 := New(200)
	d4 := New(-100)

	tests := []struct {
		d1, d2   Decimal
		expected int
	}{
		{d1, d2, 0},
		{d3, d1, 1},
		{d4, d1, -1},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := tt.d1.Cmp(tt.d2)
			if result != tt.expected {
				t.Errorf("Cmp() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestZero(t *testing.T) {
	tests := []struct {
		d        Decimal
		expected bool
	}{
		{ZERO, true},
		{New(0), true},
		{New(0.0), true},
		{New(-0.0), true},
		{New(1), false},
		{New(-1), false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if tt.d.Zero() != tt.expected {
				t.Errorf("%v.Zero() = %v, want %v", tt.d, tt.d.Zero(), tt.expected)
			}
		})
	}
}

func TestFloat(t *testing.T) {
	tests := []struct {
		d        Decimal
		expected float64
	}{
		{New(0), 0},
		{New(123.456), 123.456},
		{New(-789.012), -789.012},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if tt.d.Float() != tt.expected {
				t.Errorf("%v.Float() = %v, want %v", tt.d, tt.d.Float(), tt.expected)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		d        Decimal
		expected string
	}{
		{New(0), "0"},
		{New(123.456), "123.456"},
		{New(-789.012), "-789.012"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if tt.d.String() != tt.expected {
				t.Errorf("%v.String() = %s, want %s", tt.d, tt.d.String(), tt.expected)
			}
		})
	}
}

func TestFormattedString(t *testing.T) {
	d := New(123.456789)
	if d.FormattedString(2) != "123.46" {
		t.Errorf("FormattedString(2) = %s, want \"123.46\"", d.FormattedString(2))
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		d        Decimal
		expected float64
	}{
		{New(123), 123},
		{New(-123), 123},
		{New(0), 0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := tt.d.Abs()
			if result.Float() != tt.expected {
				t.Errorf("%v.Abs() = %v, want %v", tt.d, result.Float(), tt.expected)
			}
		})
	}
}

func TestNeg(t *testing.T) {
	tests := []struct {
		d        Decimal
		expected float64
	}{
		{New(123), -123},
		{New(-123), 123},
		{New(0), 0},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := tt.d.Neg()
			if result.Float() != tt.expected {
				t.Errorf("%v.Neg() = %v, want %v", tt.d, result.Float(), tt.expected)
			}
		})
	}
}

func TestMax(t *testing.T) {
	d1 := New(100)
	d2 := New(200)
	d3 := New(150)

	tests := []struct {
		d1, d2   Decimal
		expected float64
	}{
		{d1, d2, 200},
		{d2, d1, 200},
		{d3, d3, 150},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := tt.d1.Max(tt.d2)
			if result.Float() != tt.expected {
				t.Errorf("Max() = %v, want %v", result.Float(), tt.expected)
			}
		})
	}
}

func TestMin(t *testing.T) {
	d1 := New(100)
	d2 := New(200)
	d3 := New(150)

	tests := []struct {
		d1, d2   Decimal
		expected float64
	}{
		{d1, d2, 100},
		{d2, d1, 100},
		{d3, d3, 150},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := tt.d1.Min(tt.d2)
			if result.Float() != tt.expected {
				t.Errorf("Min() = %v, want %v", result.Float(), tt.expected)
			}
		})
	}
}

func TestSqrt(t *testing.T) {
	tests := []struct {
		d        Decimal
		expected float64
	}{
		{New(0), 0},
		{New(1), 1},
		{New(4), 2},
		{New(100), 10},
		{New(2.25), 1.5},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := tt.d.Sqrt()
			if math.Abs(result.Float()-tt.expected) > 1e-10 {
				t.Errorf("%v.Sqrt() = %v, want %v", tt.d, result.Float(), tt.expected)
			}
		})
	}
}

func TestPow(t *testing.T) {
	tests := []struct {
		d        Decimal
		y        int
		expected float64
	}{
		{New(2), 0, 1},
		{New(2), 1, 2},
		{New(2), 2, 4},
		{New(2), 3, 8},
		{New(2), 10, 1024},
		{New(2), -1, 0.5},
		{New(2), -2, 0.25},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := tt.d.Pow(tt.y)
			if math.Abs(result.Float()-tt.expected) > 1e-10 {
				t.Errorf("%v.Pow(%d) = %v, want %v", tt.d, tt.y, result.Float(), tt.expected)
			}
		})
	}
}

func TestSign(t *testing.T) {
	tests := []struct {
		d        Decimal
		expected int
	}{
		{New(-10), -1},
		{New(0), 0},
		{New(10), 1},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if tt.d.Sign() != tt.expected {
				t.Errorf("%v.Sign() = %d, want %d", tt.d, tt.d.Sign(), tt.expected)
			}
		})
	}
}

func TestIsNegative(t *testing.T) {
	tests := []struct {
		d        Decimal
		expected bool
	}{
		{New(-10), true},
		{New(0), false},
		{New(10), false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if tt.d.IsNegative() != tt.expected {
				t.Errorf("%v.IsNegative() = %v, want %v", tt.d, tt.d.IsNegative(), tt.expected)
			}
		})
	}
}

func TestIsPositive(t *testing.T) {
	tests := []struct {
		d        Decimal
		expected bool
	}{
		{New(-10), false},
		{New(0), false},
		{New(10), true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if tt.d.IsPositive() != tt.expected {
				t.Errorf("%v.IsPositive() = %v, want %v", tt.d, tt.d.IsPositive(), tt.expected)
			}
		})
	}
}

func TestIsZero(t *testing.T) {
	tests := []struct {
		d        Decimal
		expected bool
	}{
		{New(-10), false},
		{New(0), true},
		{New(10), false},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			if tt.d.IsZero() != tt.expected {
				t.Errorf("%v.IsZero() = %v, want %v", tt.d, tt.d.IsZero(), tt.expected)
			}
		})
	}
}

func TestDecimalConstants(t *testing.T) {
	if !ZERO.IsZero() {
		t.Error("ZERO should be zero")
	}
	if !ONE.EQ(New(1)) {
		t.Error("ONE should equal 1")
	}
}

// Benchmark tests
func BenchmarkDecimalAdd(b *testing.B) {
	d1 := New(123.456)
	d2 := New(78.9)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = d1.Add(d2)
	}
}

func BenchmarkDecimalMul(b *testing.B) {
	d1 := New(123.456)
	d2 := New(78.9)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = d1.Mul(d2)
	}
}

func BenchmarkDecimalDiv(b *testing.B) {
	d1 := New(123.456)
	d2 := New(78.9)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = d1.Div(d2)
	}
}

func BenchmarkDecimalSqrt(b *testing.B) {
	d := New(123.456)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = d.Sqrt()
	}
}

func BenchmarkDecimalPow(b *testing.B) {
	d := New(123.456)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = d.Pow(10)
	}
}

func TestRound(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.4", "1"},
		{"1.5", "2"},
		{"1.6", "2"},
		{"-1.4", "-1"},
		{"-1.5", "-2"},
		{"-1.6", "-2"},
		{"0.5", "1"},
		{"-0.5", "-1"},
		{"2.5", "3"},
		{"-2.5", "-3"},
		{"0", "0"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			d, _ := NewFromStringWithError(tt.input)
			result := d.Round()
			if result.String() != tt.expected {
				t.Errorf("Round() = %s, want %s", result.String(), tt.expected)
			}
		})
	}
}

func TestFloor(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.4", "1"},
		{"1.9", "1"},
		{"-1.4", "-2"},
		{"-1.9", "-2"},
		{"1", "1"},
		{"-1", "-1"},
		{"0", "0"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			d, _ := NewFromStringWithError(tt.input)
			result := d.Floor()
			if result.String() != tt.expected {
				t.Errorf("Floor() = %s, want %s", result.String(), tt.expected)
			}
		})
	}
}

func TestCeil(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.1", "2"},
		{"1.9", "2"},
		{"-1.1", "-1"},
		{"-1.9", "-1"},
		{"1", "1"},
		{"-1", "-1"},
		{"0", "0"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			d, _ := NewFromStringWithError(tt.input)
			result := d.Ceil()
			if result.String() != tt.expected {
				t.Errorf("Ceil() = %s, want %s", result.String(), tt.expected)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.9", "1"},
		{"-1.9", "-1"},
		{"1", "1"},
		{"-1", "-1"},
		{"0", "0"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			d, _ := NewFromStringWithError(tt.input)
			result := d.Truncate()
			if result.String() != tt.expected {
				t.Errorf("Truncate() = %s, want %s", result.String(), tt.expected)
			}
		})
	}
}

func TestFrac(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.5", "0.5"},
		{"-1.5", "-0.5"},
		{"1.999", "0.999"},
		{"1", "0"},
		{"0", "0"},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			d, _ := NewFromStringWithError(tt.input)
			result := d.Frac()
			if result.String() != tt.expected {
				t.Errorf("Frac() = %s, want %s", result.String(), tt.expected)
			}
		})
	}
}

func TestDecimal_Concurrency(t *testing.T) {
	a := New(100)
	b := New(50)

	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func() {
			a.Add(b)
			a.Sub(b)
			a.Mul(b)
			a.Div(b)
			done <- true
		}()
	}

	for i := 0; i < 100; i++ {
		<-done
	}
}
