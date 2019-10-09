package funcs

import "libs.altipla.consulting/money"

func Price(value int32) string {
	return money.NewFromCents(int64(value)).Format(2)
}

func Money(currency string, value int32) string {
	return money.NewFromCents(int64(value)).Display(currency)
}
