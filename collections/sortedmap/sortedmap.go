package sortedmap

import (
	"github.com/google/btree"
)

type ItemIterator func(i Item) bool

type Item interface {
	// Less tests whether the current item is less than the given argument.
	//
	// This must provide a strict weak ordering.
	// If !a.Less(b) && !b.Less(a), we treat this to mean a == b (i.e. we can only
	// hold one of either a or b in the tree).
	Less(than Item) bool

	// Key returns the key of the item.
	Key() string
}

type twrapper struct {
	item Item
}

func (t twrapper) Less(than btree.Item) bool {
	return t.item.Less(than.(twrapper).item)
}

type SortedMap struct {
	t    *btree.BTree
	vals map[string]Item
}

func New() *SortedMap {
	return &SortedMap{
		t:    btree.New(32),
		vals: make(map[string]Item),
	}
}

func (m *SortedMap) Insert(task Item) {
	m.Delete(task.Key())
	m.t.ReplaceOrInsert(twrapper{task})
	m.vals[task.Key()] = task

	if m.t.Len() != len(m.vals) {
		panic("sorted map lengths should be the same")
	}
}

func (m *SortedMap) Delete(key string) {
	v, ok := m.vals[key]
	if !ok {
		return
	}

	delete(m.vals, key)
	m.t.Delete(twrapper{v})
}

func (m *SortedMap) Ascend(fn ItemIterator) {
	m.t.Ascend(func(item btree.Item) bool {
		return fn(item.(twrapper).item)
	})
}

func (m *SortedMap) Len() int {
	return m.t.Len()
}

func (m *SortedMap) Clear() {
	m.t.Clear(false)
	m.vals = make(map[string]Item)
}

func (m *SortedMap) Min() Item {
	return m.t.Min().(twrapper).item
}
