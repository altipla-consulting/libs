package funcs

import "libs.altipla.consulting/money"

func Price(value int32) string {
	return money.FromCents(value).Format(money.FormatConfig{})
}

func Money(currency string, value int32) string {
	return money.FromCents(value).Format(money.Currency(currency))
}
