package money

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCents(t *testing.T) {
	money := NewFromCents(12579)

	require.EqualValues(t, money.Cents(), 12579)
}

func TestParse(t *testing.T) {
	money, err := Parse("125,79")
	require.NoError(t, err)

	require.EqualValues(t, money.Cents(), 12579)
}

func TestParseMultipleDecimals(t *testing.T) {
	money, err := Parse("125.7923")
	require.NoError(t, err)

	require.EqualValues(t, money.Cents(), 12579)
}

func TestParseMultipleDecimalsNoRound(t *testing.T) {
	money, err := Parse("125.7963")
	require.NoError(t, err)

	require.EqualValues(t, money.Cents(), 12579)
}

func TestFormatPrecisionFour(t *testing.T) {
	money := NewFromCents(12345)

	require.EqualValues(t, money.Format(4), "123.4500")
}

func TestFormatPrecisionThree(t *testing.T) {
	money := NewFromCents(12345)

	require.EqualValues(t, money.Format(3), "123.450")
}

func TestFormatPrecisionTwo(t *testing.T) {
	money := NewFromCents(12345)

	require.EqualValues(t, money.Format(2), "123.45")
}

func TestFormatPrecisionOne(t *testing.T) {
	money := NewFromCents(12346)

	require.EqualValues(t, money.Format(1), "123.5")
}

func TestFormatPrecisionZero(t *testing.T) {
	money := NewFromCents(12346)

	require.EqualValues(t, money.Format(0), "123")
}

func TestMul(t *testing.T) {
	money := NewFromCents(1000)

	require.EqualValues(t, money.Mul(3).Cents(), 3000)
}

func TestAdd(t *testing.T) {
	money := NewFromCents(10000)

	require.EqualValues(t, money.Add(NewFromCents(5000)).Cents(), 15000)
}

func TestSub(t *testing.T) {
	money := NewFromCents(10000)

	require.EqualValues(t, money.Sub(NewFromCents(4000)).Cents(), 6000)
}

func TestDiv(t *testing.T) {
	money := NewFromCents(1234)

	require.EqualValues(t, money.Div(NewFromCents(2)).Cents(), int64(617))
}

func TestLessThan(t *testing.T) {
	money := NewFromCents(10000)
	other := NewFromCents(5000)

	require.False(t, money.LessThan(other))
	require.True(t, other.LessThan(money))
}

func TestAddTaxPercent(t *testing.T) {
	money := NewFromCents(10000)

	require.EqualValues(t, money.AddTaxPercent(20).Cents(), 12000)
}

func TestIsZero(t *testing.T) {
	money := New()
	require.True(t, money.IsZero())

	money = money.Add(NewFromCents(1))
	require.False(t, money.IsZero())
}

func TestDisplayEUR(t *testing.T) {
	money := NewFromCents(12345)

	require.Equal(t, money.Display("EUR"), "123.45 €")
}

func TestDisplayUSD(t *testing.T) {
	money := NewFromCents(12345)

	require.Equal(t, money.Display("USD"), "$123.45")
}
