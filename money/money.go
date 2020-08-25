package money

import (
	"fmt"
	"strconv"
	"strings"

	"libs.altipla.consulting/errors"
)

var (
	EUR = FormatConfig{
		Symbol:   "\u20ac",
		Thousand: ",",
	}

	USD = FormatConfig{
		Symbol:   "$",
		Prefix:   true,
		Thousand: ",",
	}

	GBP = FormatConfig{
		Symbol:   "\u00a3",
		Prefix:   true,
		Thousand: ",",
	}

	MXN = FormatConfig{
		Symbol:   "Mex$",
		Prefix:   true,
		Thousand: ",",
	}
)

type FormatConfig struct {
	Symbol        string
	Prefix        bool
	Thousand      string
	Decimal       string
	ForceDecimals bool
}

// Currency returns the configuration of a currency by its name. Unknown currencies will return the default config.
func Currency(currency string) FormatConfig {
	currencies := map[string]FormatConfig{
		"EUR": EUR,
		"USD": USD,
		"GBP": GBP,
		"MXN": MXN,
	}

	return currencies[currency]
}

type Money int32

// FromCents creates a new instance with a cents value.
func FromCents(cents int32) Money {
	return Money(cents)
}

// Parse a string to create a new money value. It can read `XX.YY` and `XX,YY`.
// An empty string is parsed as zero.
func Parse(s string) (Money, error) {
	if len(s) == 0 {
		return Money(0), nil
	}

	s = strings.Replace(s, ",", ".", -1)

	var amount int32
	switch parts := strings.Split(s, "."); len(parts) {
	case 1:
		var err error
		units, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return Money(0), errors.Trace(err)
		}
		amount = int32(units) * 100

	case 2:
		units, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return Money(0), errors.Trace(err)
		}
		if len(parts[1]) > 2 {
			parts[1] = parts[1][:2]
		}
		decimals, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return Money(0), errors.Trace(err)
		}
		if len(parts[1]) == 1 {
			amount = int32(units)*100 + int32(decimals)*10
		} else {
			amount = int32(units)*100 + int32(decimals)
		}

	default:
		return Money(0), errors.Errorf("cannot parse money value: %v", s)
	}

	return Money(amount), nil
}

// Format the money value with a specific configuration.
func (m Money) Format(config FormatConfig) string {
	if config.Decimal == "" {
		config.Decimal = "."
	}

	value := strconv.FormatInt(int64(m), 10)
	value = fmt.Sprintf("%03v", value)
	if config.Thousand != "" {
		for i := len(value) - 5; i > 0; i -= 3 {
			value = value[:i] + config.Thousand + value[i:]
		}
	}

	value = value[:len(value)-2] + config.Decimal + value[len(value)-2:]

	if config.Symbol != "" {
		if config.Prefix {
			value = config.Symbol + value
		} else {
			value = value + " " + config.Symbol
		}
	}

	if !config.ForceDecimals {
		value = strings.TrimSuffix(value, config.Decimal+"00")
	}
	return value
}

// Format the money value with the default configuration
func (m Money) String() string {
	return m.Format(FormatConfig{})
}

// Cents returns the value with cents precision (2 decimal places) as a number.
func (m Money) Cents() int32 {
	return int32(m)
}

// IsZero returns true if there is no money.
func (m Money) IsZero() bool {
	return m == 0
}

// LessThan returns true if a money value is less than the other.
func (m Money) LessThan(other Money) bool {
	return m < other
}

// Mul multiplies the money value n times and returns the result.
func (m Money) Mul(n int32) Money {
	return Money(int32(m) * n)
}

// Add two money values together and returns the result.
func (m Money) Add(other Money) Money {
	return Money(int32(m) + int32(other))
}

// Sub subtracts two money values and returns the result.
func (m Money) Sub(other Money) Money {
	return Money(int32(m) - int32(other))
}

// Div divides two money values and returns the result.
func (m Money) Div(other Money) Money {
	return Money(int32(m) / int32(other))
}

// AddTaxPercent adds a percentage of the price to itself.
func (m Money) AddTaxPercent(tax int32) Money {
	return m.Mul(tax).Div(Money(100)).Add(m)
}
