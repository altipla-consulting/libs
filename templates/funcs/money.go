package funcs

import "github.com/altipla-consulting/money"

func Price(value int32) string {
	return money.NewFromCents(int64(value)).Format(2)
}
