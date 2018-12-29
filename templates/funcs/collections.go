package funcs

import (
	"math/rand"
	"reflect"
)

func GenRange(n, start int64) []int64 {
	nums := make([]int64, n)
	for i := range nums {
		nums[i] = int64(i) + start
	}

	return nums
}

func Shuffle(slice interface{}) interface{} {
	if slice == nil {
		return nil
	}

	value := reflect.ValueOf(slice)
	perm := rand.Perm(value.Len())

	result := reflect.MakeSlice(reflect.TypeOf(slice), value.Len(), value.Len())
	for i, idx := range perm {
		result.Index(i).Set(value.Index(idx))
	}

	return result.Interface()
}

func Limit(max int, slice interface{}) interface{} {
	if slice == nil {
		return nil
	}

	value := reflect.ValueOf(slice)
	if value.Len() <= max {
		return value.Interface()
	}

	result := reflect.MakeSlice(reflect.TypeOf(slice), max, max)
	for i := 0; i < max; i++ {
		result.Index(i).Set(value.Index(i))
	}

	return result.Interface()
}

func Slice(min, max int, slice interface{}) interface{} {
	if slice == nil {
		return nil
	}

	value := reflect.ValueOf(slice)
	if value.Len() <= min {
		return nil
	}

	if max > value.Len() {
		max = value.Len()
	}

	size := max - min
	result := reflect.MakeSlice(reflect.TypeOf(slice), size, size)
	var current int
	for i := min; i < max; i++ {
		result.Index(current).Set(value.Index(i))
		current++
	}

	return result.Interface()
}

func RandItem(slice interface{}) interface{} {
	if slice == nil {
		return nil
	}

	value := reflect.ValueOf(slice)
	if value.Len() == 0 {
		return nil
	}

	return value.Index(rand.Intn(value.Len())).Interface()
}

func Last(index int, list interface{}) bool {
	return reflect.ValueOf(list).Len()-1 == index
}
