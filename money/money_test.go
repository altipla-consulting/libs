package money

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	money, err := Parse("125,79")
	require.NoError(t, err)

	require.EqualValues(t, money, 12579)
}

func TestParseMultipleDecimals(t *testing.T) {
	money, err := Parse("125.7923")
	require.NoError(t, err)

	require.EqualValues(t, money, 12579)
}

func TestParseMultipleDecimalsNoRound(t *testing.T) {
	money, err := Parse("125.7963")
	require.NoError(t, err)

	require.EqualValues(t, money, 12579)
}

func TestParseFloatError(t *testing.T) {
	money, err := Parse("10.03")
	require.NoError(t, err)

	require.EqualValues(t, money, 1003)
}

func TestParseOneDecimal(t *testing.T) {
	money, err := Parse("10.3")
	require.NoError(t, err)

	require.EqualValues(t, money, 1030)
}

func TestParseOneDecimalWithZero(t *testing.T) {
	money, err := Parse("10.30")
	require.NoError(t, err)

	require.EqualValues(t, money, 1030)
}

func TestParseFloatErrorIterative(t *testing.T) {
	for i := 1000; i <= 9999; i++ {
		var prefix string
		if i%100 < 10 {
			prefix = "0"
		}
		s := fmt.Sprintf("%v.%s%v", i/100, prefix, i%100)
		money, err := Parse(s)
		require.NoError(t, err)

		require.EqualValues(t, money, i, "i: %v; s: %v", i, s)
	}
}

func TestParseWithoutDecimals(t *testing.T) {
	money, err := Parse("10")
	require.NoError(t, err)

	require.EqualValues(t, money, 1000)
}

func TestFormatDefault(t *testing.T) {
	money := Money(12345)

	require.EqualValues(t, money.Format(FormatConfig{}), "123.45")
}

func TestFormatSymbol(t *testing.T) {
	money := Money(12345)

	require.EqualValues(t, money.Format(EUR), "123.45 €")
}

func TestFormatSymbolPrefix(t *testing.T) {
	money := Money(12345)

	require.EqualValues(t, money.Format(USD), "$123.45")
}

func TestFormatNegativeSymbolPrefix(t *testing.T) {
	money := Money(-12345)

	cnf := FormatConfig{
		Symbol: "$",
		Prefix: true,
	}
	require.EqualValues(t, money.Format(cnf), "$-123.45")
}

func TestFormatThousand(t *testing.T) {
	money := Money(123456)

	require.EqualValues(t, money.Format(EUR), "1,234.56 €")
}

func TestFormatCents(t *testing.T) {
	money := Money(12)

	require.EqualValues(t, money.Format(FormatConfig{}), "0.12")
}

func TestFormatZeroDecimals(t *testing.T) {
	money := Money(100)

	require.EqualValues(t, money.Format(FormatConfig{}), "1")
}

func TestFormatForceDecimals(t *testing.T) {
	money := Money(100)

	cnf := FormatConfig{
		ForceDecimals: true,
	}
	require.EqualValues(t, money.Format(cnf), "1.00")
}

func TestFormatCurrencyEUR(t *testing.T) {
	money := Money(123456)

	require.EqualValues(t, money.Format(Currency("EUR")), "1,234.56 €")
}

func TestFormatCurrencyUSD(t *testing.T) {
	money := Money(123456)

	require.EqualValues(t, money.Format(Currency("USD")), "$1,234.56")
}

func TestFormatCurrencyGBP(t *testing.T) {
	money := Money(123456)

	require.EqualValues(t, money.Format(Currency("GBP")), "£1,234.56")
}

func TestFormatCurrencyMXN(t *testing.T) {
	money := Money(123456)

	require.EqualValues(t, money.Format(Currency("MXN")), "Mex$1,234.56")
}

func TestFormatEmptyCurrency(t *testing.T) {
	money := Money(123456)

	require.EqualValues(t, money.Format(Currency("")), "1234.56")
}

func TestFormatUnknownCurrency(t *testing.T) {
	money := Money(123456)

	require.EqualValues(t, money.Format(Currency("PLN")), "1234.56")
}

func TestLessThan(t *testing.T) {
	money := Money(10000)
	other := Money(5000)

	require.False(t, money.LessThan(other))
	require.True(t, other.LessThan(money))
}

func TestIsZero(t *testing.T) {
	money := Money(0)
	require.True(t, money.IsZero())

	money = Money(1)
	require.False(t, money.IsZero())
}

func TestMul(t *testing.T) {
	money := Money(1000)

	require.EqualValues(t, money.Mul(3), 3000)
}

func TestAdd(t *testing.T) {
	money := Money(10000)

	require.EqualValues(t, money.Add(Money(5000)), 15000)
}

func TestSub(t *testing.T) {
	money := Money(10000)

	require.EqualValues(t, money.Sub(Money(4000)), 6000)
}

func TestDiv(t *testing.T) {
	money := Money(1234)

	require.EqualValues(t, money.Div(Money(2)), int32(617))
}

func TestAddTaxPercent(t *testing.T) {
	money := Money(10000)

	require.EqualValues(t, money.AddTaxPercent(20), 12000)
}
