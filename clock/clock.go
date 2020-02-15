package clock

import (
	"time"
)

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (c realClock) Now() time.Time {
	return time.Now()
}

func New() Clock {
	return realClock{}
}

type staticClock struct {
	t time.Time
}

func (c staticClock) Now() time.Time {
	return c.t
}

func NewStatic(t time.Time) Clock {
	return staticClock{t}
}
