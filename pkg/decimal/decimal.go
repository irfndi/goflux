package decimal

import (
	"math/big"
)

const defaultPrecision = 256

func newFloat() *big.Float {
	return new(big.Float).SetPrec(defaultPrecision).SetMode(big.ToNearestEven)
}

func (d Decimal) valueOrZero() *big.Float {
	if d.val == nil {
		return newFloat().SetInt64(0)
	}
	return d.val
}

// Decimal represents a high-precision decimal number.
// It wraps math/big.Float to provide convenient methods for financial calculations.
type Decimal struct {
	val *big.Float
}

var (
	// ZERO is a Decimal with value 0
	ZERO = New(0)
	// ONE is a Decimal with value 1
	ONE = New(1)
)

// New creates a new Decimal from a float64
func New(f float64) Decimal {
	return Decimal{val: newFloat().SetFloat64(f)}
}

// NewFromInt creates a new Decimal from an int64
func NewFromInt(i int64) Decimal {
	return Decimal{val: newFloat().SetInt64(i)}
}

// NewFromString creates a new Decimal from a string.
// It panics if string is not a valid number.
func NewFromString(s string) Decimal {
	val, _, err := big.ParseFloat(s, 10, 256, big.ToNearestEven)
	if err != nil {
		panic("invalid decimal string: " + s)
	}
	return Decimal{val: val}
}

// NewFromStringWithError creates a new Decimal from a string.
// It returns an error if the string is not a valid number.
func NewFromStringWithError(s string) (Decimal, error) {
	val, _, err := big.ParseFloat(s, 10, 256, big.ToNearestEven)
	if err != nil {
		return Decimal{}, err
	}
	return Decimal{val: val}, nil
}

// NewFromBigFloat creates a new Decimal from a big.Float
func NewFromBigFloat(f *big.Float) Decimal {
	if f == nil {
		return Decimal{val: newFloat().SetInt64(0)}
	}
	return Decimal{val: new(big.Float).Copy(f)}
}

// Add returns d + d2
func (d Decimal) Add(d2 Decimal) Decimal {
	return Decimal{val: newFloat().Add(d.valueOrZero(), d2.valueOrZero())}
}

// Sub returns d - d2
func (d Decimal) Sub(d2 Decimal) Decimal {
	return Decimal{val: newFloat().Sub(d.valueOrZero(), d2.valueOrZero())}
}

// Mul returns d * d2
func (d Decimal) Mul(d2 Decimal) Decimal {
	return Decimal{val: newFloat().Mul(d.valueOrZero(), d2.valueOrZero())}
}

// Div returns d / d2
func (d Decimal) Div(d2 Decimal) Decimal {
	if d2.valueOrZero().Sign() == 0 {
		return ZERO
	}
	return Decimal{val: newFloat().Quo(d.valueOrZero(), d2.valueOrZero())}
}

// GT returns true if d > d2
func (d Decimal) GT(d2 Decimal) bool {
	return d.valueOrZero().Cmp(d2.valueOrZero()) > 0
}

// GTE returns true if d >= d2
func (d Decimal) GTE(d2 Decimal) bool {
	return d.valueOrZero().Cmp(d2.valueOrZero()) >= 0
}

// LT returns true if d < d2
func (d Decimal) LT(d2 Decimal) bool {
	return d.valueOrZero().Cmp(d2.valueOrZero()) < 0
}

// LTE returns true if d <= d2
func (d Decimal) LTE(d2 Decimal) bool {
	return d.valueOrZero().Cmp(d2.valueOrZero()) <= 0
}

// EQ returns true if d == d2
func (d Decimal) EQ(d2 Decimal) bool {
	return d.valueOrZero().Cmp(d2.valueOrZero()) == 0
}

// Zero returns true if d == 0
func (d Decimal) Zero() bool {
	return d.valueOrZero().Sign() == 0
}

// Float returns float64 representation of d
func (d Decimal) Float() float64 {
	f, _ := d.valueOrZero().Float64()
	return f
}

// String returns string representation of d
func (d Decimal) String() string {
	return d.valueOrZero().Text('f', -1)
}

// FormattedString returns string representation of d with fixed precision
func (d Decimal) FormattedString(precision int) string {
	return d.valueOrZero().Text('f', precision)
}

// Abs returns absolute value of d
func (d Decimal) Abs() Decimal {
	return Decimal{val: newFloat().Abs(d.valueOrZero())}
}

// Neg returns -d
func (d Decimal) Neg() Decimal {
	return Decimal{val: newFloat().Neg(d.valueOrZero())}
}

// Max returns larger of d and d2
func (d Decimal) Max(d2 Decimal) Decimal {
	if d.GT(d2) {
		return d
	}
	return d2
}

// Min returns smaller of d and d2
func (d Decimal) Min(d2 Decimal) Decimal {
	if d.LT(d2) {
		return d
	}
	return d2
}

// Sqrt returns square root of d
func (d Decimal) Sqrt() Decimal {
	return Decimal{val: newFloat().Sqrt(d.valueOrZero())}
}

// Pow returns d^y where y is an integer
func (d Decimal) Pow(y int) Decimal {
	if y == 0 {
		return ONE
	}

	absY := y
	neg := false
	if y < 0 {
		absY = -y
		neg = true
	}

	result := ONE
	base := d
	for absY > 0 {
		if absY&1 == 1 {
			result = result.Mul(base)
		}
		base = base.Mul(base)
		absY >>= 1
	}

	if neg {
		return ONE.Div(result)
	}
	return result
}

// Cmp compares d and d2 and returns:
//
//	-1 if d <  d2
//	 0 if d == d2
//	+1 if d >  d2
func (d Decimal) Cmp(d2 Decimal) int {
	return d.valueOrZero().Cmp(d2.valueOrZero())
}

// Sign returns -1 if d < 0, 0 if d == 0, +1 if d > 0
func (d Decimal) Sign() int {
	return d.valueOrZero().Sign()
}

// IsNegative returns true if d < 0
func (d Decimal) IsNegative() bool {
	return d.Sign() < 0
}

// IsPositive returns true if d > 0
func (d Decimal) IsPositive() bool {
	return d.Sign() > 0
}

// IsZero returns true if d == 0
func (d Decimal) IsZero() bool {
	return d.Sign() == 0
}

// Round returns d rounded to the nearest integer
func (d Decimal) Round() Decimal {
	intPart := new(big.Int)
	d.valueOrZero().Int(intPart)

	intPartFloat := new(big.Float).SetInt(intPart)

	frac := new(big.Float).Sub(d.valueOrZero(), intPartFloat)
	half := new(big.Float).SetFloat64(0.5)

	if d.valueOrZero().Sign() >= 0 {
		if frac.Cmp(half) >= 0 {
			intPart.Add(intPart, new(big.Int).SetInt64(1))
		}
	} else {
		negHalf := new(big.Float).Neg(half)
		if frac.Cmp(negHalf) <= 0 {
			intPart.Sub(intPart, new(big.Int).SetInt64(1))
		}
	}

	result := new(big.Float).SetInt(intPart)
	return Decimal{val: newFloat().Set(result)}
}

// Floor returns greatest integer value less than or equal to d
func (d Decimal) Floor() Decimal {
	result := new(big.Int)
	d.valueOrZero().Int(result)

	resultFloat := new(big.Float).SetInt(result)

	if d.IsNegative() && d.valueOrZero().Cmp(resultFloat) != 0 {
		result.Add(result, new(big.Int).SetInt64(-1))
	}

	return Decimal{val: newFloat().SetInt(result)}
}

// Ceil returns least integer value greater than or equal to d
func (d Decimal) Ceil() Decimal {
	result := new(big.Int)
	d.valueOrZero().Int(result)

	resultFloat := new(big.Float).SetInt(result)

	if d.IsPositive() && d.valueOrZero().Cmp(resultFloat) != 0 {
		result.Add(result, new(big.Int).SetInt64(1))
	}

	return Decimal{val: newFloat().SetInt(result)}
}

// Truncate returns integer part of d, dropping any fractional part
func (d Decimal) Truncate() Decimal {
	result := new(big.Int)
	d.valueOrZero().Int(result)
	return Decimal{val: newFloat().SetInt(result)}
}

// Frac returns the fractional part of d
func (d Decimal) Frac() Decimal {
	intPart := new(big.Int)
	_, _ = d.valueOrZero().Int(intPart)
	intPartFloat := new(big.Float).SetInt(intPart)
	frac := new(big.Float).Sub(d.valueOrZero(), intPartFloat)
	return Decimal{val: newFloat().Set(frac)}
}
