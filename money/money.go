package money

import (
	"math"
	"strconv"
	"strings"

	"github.com/Rhymond/go-money"
	"libs.altipla.consulting/errors"
)

func init() {
	money.AddCurrency("EUR", "\u20ac", "1 $", ".", ",", 2)
}

// Money represents a monetary value
type Money struct {
	value *money.Money
}

// NewFromCents creates a new instance with a cents value.
func NewFromCents(cents int64) *Money {
	return &Money{money.New(cents, "")}
}

// New creates a new instance with a zero value.
func New() *Money {
	return NewFromCents(0)
}

// Parse a string to create a new money value. It can read `XX.YY` and `XX,YY`.
// An empty string is parsed as zero.
func Parse(s string) (*Money, error) {
	if len(s) == 0 {
		return New(), nil
	}

	s = strings.Replace(s, ",", ".", -1)

	var amount int64
	switch parts := strings.Split(s, "."); len(parts) {
	case 1:
		var err error
		units, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return nil, errors.Trace(err)
		}
		amount = units * 100

	case 2:
		units, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return nil, errors.Trace(err)
		}
		if len(parts[1]) > 2 {
			parts[1] = parts[1][:2]
		}
		decimals, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, errors.Trace(err)
		}
		if len(parts[1]) == 1 {
			amount = units*100 + decimals*10
		} else {
			amount = units*100 + decimals
		}

	default:
		return nil, errors.Errorf("cannot parse money value: %v", s)
	}

	return NewFromCents(amount), nil
}

// Cents returns the value with cents precision (2 decimal places) as a number.
func (m *Money) Cents() int64 {
	return m.value.Amount()
}

// Format the money value with a specific decimal precision.
func (m *Money) Format(prec int) string {
	value := m.value.Amount()
	switch {
	case prec > 2:
		value *= int64(math.Pow10(prec - 2))

	case prec == 1:
		if value%10 > 5 {
			value = (value / 10) + 1
		} else {
			value = value / 10
		}

	case prec == 0:
		value = value / 100
	}

	cur := money.GetCurrency("EUR")
	f := money.NewFormatter(prec, cur.Decimal, cur.Thousand, cur.Grapheme, "1")
	return f.Format(value)
}

// Mul multiplies the money value n times and returns the result.
func (m *Money) Mul(n int64) *Money {
	return &Money{m.value.Multiply(n)}
}

// Add two money values together and returns the result.
func (m *Money) Add(other *Money) *Money {
	result, err := m.value.Add(other.value)
	if err != nil {
		panic(err)
	}

	return &Money{result}
}

// Sub subtracts two money values and returns the result.
func (m *Money) Sub(other *Money) *Money {
	result, err := m.value.Subtract(other.value)
	if err != nil {
		panic(err)
	}

	return &Money{result}
}

// Div divides two money values and returns the result.
func (m *Money) Div(other *Money) *Money {
	return &Money{m.value.Divide(other.value.Amount())}
}

// LessThan returns true if a money value is less than the other.
func (m *Money) LessThan(other *Money) bool {
	result, err := m.value.LessThan(other.value)
	if err != nil {
		panic(err)
	}

	return result
}

// AddTaxPercent adds a percentage of the price to itself.
func (m *Money) AddTaxPercent(tax int64) *Money {
	result, err := m.value.Multiply(tax).Divide(100).Add(m.value)
	if err != nil {
		panic(err)
	}

	return &Money{result}
}

// IsZero returns true if there is no money.
func (m *Money) IsZero() bool {
	return m.value.IsZero()
}

// Display formats the money value with a specific currency.
func (m *Money) Display(currency string) string {
	return money.GetCurrency(currency).Formatter().Format(m.value.Amount())
}
