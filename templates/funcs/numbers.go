package funcs

import "fmt"

func Even(n int) bool {
	return n%2 == 0
}

func Odd(n int) bool {
	return n%2 == 1
}

func Trio(n int) bool {
	return (n+1)%3 == 0
}

func Mod(n, m int) int {
	return n % m
}

func Div(a, b int) int {
	return a / b
}

func Times(a, b int64) int64 {
	return a * b
}

func Add(a, b interface{}) (int64, error) {
	var ai, bi int64

	if n, ok := a.(int64); ok {
		ai = n
	} else if n, ok := a.(int); ok {
		ai = int64(n)
	} else {
		return 0, fmt.Errorf("invalid add first argument: %+v", a)
	}

	if n, ok := b.(int64); ok {
		bi = n
	} else if n, ok := b.(int); ok {
		bi = int64(n)
	} else {
		return 0, fmt.Errorf("invalid add second argument: %+v", b)
	}

	return ai + bi, nil
}

func Percentage(old, current int64) int64 {
	return int64(float64(old-current) / float64(old) * 100.)
}
