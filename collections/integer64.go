package collections

// HasInt64 returns true if list contains search.
func HasInt64(list []int64, search int64) bool {
	for _, item := range list {
		if item == search {
			return true
		}
	}

	return false
}

// UniqueIntegers64 returns list without any duplicates in it.
func UniqueIntegers64(list []int64) []int64 {
	index := map[int64]bool{}
	result := []int64{}
	for _, item := range list {
		if !index[item] {
			index[item] = true
			result = append(result, item)
		}
	}

	return result
}

// CompareInts64 returns true if the values of both lhs & rhs are the same.
func CompareInts64(lhs, rhs []int64) bool {
	if len(lhs) != len(rhs) {
		return false
	}
	for index := range lhs {
		if lhs[index] != rhs[index] {
			return false
		}
	}
	return true
}

// ReverseInt64 returns a reversed list
func ReverseInt64(list []int64) []int64 {
	result := []int64{}
	for i := len(list) - 1; i >= 0; i-- {
		result = append(result, list[i])
	}
	return result
}

// RemoveInt64 return list without remove.
func RemoveInt64(list []int64, remove int64) []int64 {
	result := []int64{}
	for _, item := range list {
		if item != remove {
			result = append(result, item)
		}
	}

	return result
}
