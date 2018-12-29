package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIncrementInPlace(t *testing.T) {
	counter := &windowCounter{
		slots:        make([]int64, 40),
		lastMove:     time.Date(2018, 2, 1, 15, 14, 13, 0, time.UTC),
		timeProvider: func() time.Time { return time.Date(2018, 2, 1, 15, 14, 13, 0, time.UTC) },
	}

	counter.incr(1)

	require.EqualValues(t, counter.total(), 1)
	require.Equal(t, counter.slots, extendSlice([]int64{1}))
}

func TestMultipleIncrements(t *testing.T) {
	counter := &windowCounter{
		slots:        make([]int64, 40),
		lastMove:     time.Date(2018, 2, 1, 15, 14, 13, 0, time.UTC),
		timeProvider: func() time.Time { return time.Date(2018, 2, 1, 15, 14, 13, 0, time.UTC) },
	}

	counter.incr(1)
	counter.incr(3)

	require.EqualValues(t, counter.total(), 4)
	require.Equal(t, counter.slots, extendSlice([]int64{4}))
}

func TestMovementInsideCell(t *testing.T) {
	counter := &windowCounter{
		slots:        extendSlice([]int64{1}),
		lastMove:     time.Date(2018, 2, 1, 15, 14, 13, 0, time.UTC),
		timeProvider: func() time.Time { return time.Date(2018, 2, 1, 15, 14, 27, 0, time.UTC) },
	}

	counter.incr(1)

	require.EqualValues(t, counter.total(), 2)
	require.Equal(t, counter.slots, extendSlice([]int64{2}))
}

func TestMovementNextCell(t *testing.T) {
	counter := &windowCounter{
		slots:        extendSlice([]int64{1}),
		lastMove:     time.Date(2018, 2, 1, 15, 14, 13, 0, time.UTC),
		timeProvider: func() time.Time { return time.Date(2018, 2, 1, 15, 14, 28, 0, time.UTC) },
	}

	counter.incr(1)

	require.EqualValues(t, counter.total(), 2)
	require.Equal(t, counter.slots, extendSlice([]int64{1, 1}))
}

func TestMovementThroughUsedCells(t *testing.T) {
	counter := &windowCounter{
		slots:        extendSlice([]int64{1, 5, 6, 7}),
		lastMove:     time.Date(2018, 2, 1, 15, 14, 13, 0, time.UTC),
		timeProvider: func() time.Time { return time.Date(2018, 2, 1, 15, 14, 58, 0, time.UTC) },
	}

	counter.incr(1)

	require.EqualValues(t, counter.total(), 2)
	require.Equal(t, counter.slots, extendSlice([]int64{1, 0, 0, 1}))
}

func TestMovementRing(t *testing.T) {
	counter := &windowCounter{
		slots:        extendSlice([]int64{1, 5, 6, 7}),
		lastMove:     time.Date(2018, 2, 1, 15, 14, 13, 0, time.UTC),
		timeProvider: func() time.Time { return time.Date(2018, 2, 1, 15, 14, 58, 0, time.UTC) },
		pos:          39,
	}

	counter.incr(1)

	require.EqualValues(t, counter.total(), 8)
	require.Equal(t, counter.slots, extendSlice([]int64{0, 0, 1, 7}))
}

func TestMoreTimeThanSlots(t *testing.T) {
	counter := &windowCounter{
		slots:        extendSlice([]int64{1, 5, 6, 7}),
		lastMove:     time.Date(2018, 2, 1, 15, 14, 13, 0, time.UTC),
		timeProvider: func() time.Time { return time.Date(2018, 2, 1, 15, 34, 28, 0, time.UTC) },
	}

	counter.incr(1)

	require.EqualValues(t, counter.total(), 1)
	require.Equal(t, counter.slots, extendSlice([]int64{0, 1}))
}

func TestExactTimeAsSlots(t *testing.T) {
	counter := &windowCounter{
		slots:        extendSlice([]int64{1, 5, 6, 7}),
		lastMove:     time.Date(2018, 2, 1, 15, 14, 13, 0, time.UTC),
		timeProvider: func() time.Time { return time.Date(2018, 2, 1, 15, 34, 13, 0, time.UTC) },
		pos:          3,
	}

	counter.incr(1)

	require.EqualValues(t, counter.total(), 1)
	require.Equal(t, counter.slots, extendSlice([]int64{0, 0, 0, 1}))
}

func TestTotalRemovingOld(t *testing.T) {
	counter := &windowCounter{
		slots:        extendSlice([]int64{1, 5, 6, 7, 8}),
		lastMove:     time.Date(2018, 2, 1, 15, 14, 13, 0, time.UTC),
		timeProvider: func() time.Time { return time.Date(2018, 2, 1, 15, 14, 58, 0, time.UTC) },
	}

	require.EqualValues(t, counter.total(), 9)
	require.Equal(t, counter.slots, extendSlice([]int64{1, 0, 0, 0, 8}))
}

func extendSlice(a []int64) []int64 {
	r := make([]int64, 40)
	copy(r, a)
	return r
}
