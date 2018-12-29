package money

import (
	"fmt"
	"math/big"
	"strings"
)

// Money represents a monetary value
type Money struct {
	rat *big.Rat
}

// New creates a new instance with a zero value.
func New() *Money {
	return &Money{
		rat: big.NewRat(0, 1),
	}
}

// NewFromCents creates a new instance with a cents value.
func NewFromCents(cents int64) *Money {
	return &Money{
		rat: big.NewRat(cents, 100),
	}
}

// Parse a string to create a new money value. It can read `XX.YY` and `XX,YY`.
// An empty string is parsed as zero.
func Parse(s string) (*Money, error) {
	if len(s) == 0 {
		return New(), nil
	}

	s = strings.Replace(s, ",", ".", -1)

	rat := new(big.Rat)
	if _, err := fmt.Sscan(s, rat); err != nil {
		return nil, fmt.Errorf("money: cannot scan value: %s: %s", s, err)
	}

	return &Money{rat}, nil
}

// Cents returns the value with cents precision (2 decimal places) as a number.
func (money *Money) Cents() int64 {
	cents := big.NewInt(100)

	v := new(big.Int)
	v.Mul(money.rat.Num(), cents)
	v.Quo(v, money.rat.Denom())

	return v.Int64()
}

// Format the money value with a specific decimal precision.
func (money *Money) Format(prec int) string {
	return money.rat.FloatString(prec)
}

// Mul multiplies the money value n times and returns the result.
func (money *Money) Mul(n int64) *Money {
	b := big.NewRat(n, 1)
	result := New()
	result.rat.Mul(money.rat, b)
	return result
}

// Add two money values together and returns the result.
func (money *Money) Add(other *Money) *Money {
	result := New()
	result.rat.Add(money.rat, other.rat)
	return result
}

// Sub subtracts two money values and returns the result.
func (money *Money) Sub(other *Money) *Money {
	result := New()
	result.rat.Sub(money.rat, other.rat)
	return result
}

// Div divides two money values and returns the result.
func (money *Money) Div(other *Money) *Money {
	result := New()
	result.rat.Quo(money.rat, other.rat)
	return result
}

// LessThan returns true if a money value is less than the other.
func (money *Money) LessThan(other *Money) bool {
	return money.rat.Cmp(other.rat) == -1
}

// AddTaxPercent adds a percentage of the price to itself.
func (money *Money) AddTaxPercent(tax int64) *Money {
	result := New()
	result.rat.Set(money.rat)
	ratTax := big.NewRat(tax, 100)
	result.rat.Add(result.rat, ratTax.Mul(ratTax, money.rat))
	return result
}

// IsZero returns true if there is no money.
func (money *Money) IsZero() bool {
	return money.Cents() == 0
}

// Markup adds a percentage with decimals of the price to itself. The
// percentage should be pre-multiplied by 100 to avoid floating point issues.
func (money *Money) Markup(tax int64) *Money {
	result := New()
	result.rat.Set(money.rat)
	ratTax := big.NewRat(tax, 10000)
	result.rat.Add(result.rat, ratTax.Mul(ratTax, money.rat))
	return result
}
