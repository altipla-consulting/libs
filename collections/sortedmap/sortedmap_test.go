package sortedmap

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type taskKey struct {
	code   string
	minETA time.Time
}

func (key *taskKey) Less(than Item) bool {
	other := than.(*taskKey)
	if key.minETA.Equal(other.minETA) {
		return key.code < other.code
	}
	return key.minETA.Before(other.minETA)
}

func (key *taskKey) Key() string {
	return key.code
}

func TestSortedMapOrder(t *testing.T) {
	foo := New()

	// No MinETA, order by code.
	foo.Insert(&taskKey{
		code: "foo-1",
	})
	foo.Insert(&taskKey{
		code: "foo-2",
	})
	foo.Insert(&taskKey{
		code: "foo-3",
	})

	// MinETA, should be in reverse order.
	foo.Insert(&taskKey{
		code:   "foo-4",
		minETA: time.Now().Add(5 * time.Minute),
	})
	foo.Insert(&taskKey{
		code:   "foo-5",
		minETA: time.Now(),
	})

	// Reinsert should not affect the item nor duplicate it.
	foo.Insert(&taskKey{
		code: "foo-1",
	})

	// Reinsert with a different MinETA should reorder it.
	foo.Insert(&taskKey{
		code:   "foo-3",
		minETA: time.Now().Add(10 * time.Minute),
	})

	var codes []string
	foo.Ascend(func(item Item) bool {
		key := item.(*taskKey)
		codes = append(codes, key.code)
		return true
	})

	require.Equal(t, codes, []string{"foo-1", "foo-2", "foo-5", "foo-4", "foo-3"})
}
