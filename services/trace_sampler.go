package services

import (
	"math/rand"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"go.opencensus.io/trace"
)

type windowCounter struct {
	slots        []int64
	pos          int
	lastMove     time.Time
	timeProvider func() time.Time

	lastQuota       time.Time
	nextMeasurement time.Time
}

func newWindowCounter() *windowCounter {
	return &windowCounter{
		slots:        make([]int64, 40),
		lastMove:     time.Now(),
		timeProvider: time.Now,
	}
}

func (counter *windowCounter) incr(value int64) {
	move := int(counter.timeProvider().Sub(counter.lastMove)) / int(15*time.Second)
	size := len(counter.slots)

	if move > size {
		move = move % size
		for i := 0; i < size; i++ {
			counter.slots[i] = 0
		}
		counter.lastMove = counter.timeProvider()
	}

	if move > 0 {
		for i := 0; i < move; i++ {
			counter.pos++
			if counter.pos >= size {
				counter.pos = 0
			}
			counter.slots[counter.pos] = 0
		}
		counter.lastMove = counter.timeProvider()
	}

	counter.slots[counter.pos] += value
}

func (counter *windowCounter) total() (result int64) {
	// Force cleaning of old values
	counter.incr(0)

	for i := 0; i < len(counter.slots); i++ {
		result += counter.slots[i]
	}
	return
}

type customSampler struct {
	mu       *sync.Mutex
	counters map[string]*windowCounter
}

func newCustomSampler() *customSampler {
	return &customSampler{
		mu:       new(sync.Mutex),
		counters: make(map[string]*windowCounter),
	}
}

func (sampler *customSampler) Sampler() trace.Sampler {
	return func(params trace.SamplingParameters) trace.SamplingDecision {
		// Do not trace requests to the profiler that happen in the background.
		if strings.HasPrefix(params.Name, "google.devtools.cloudprofiler.") {
			return trace.SamplingDecision{}
		}

		// If a parent decides to log we send all the tracing through the services.
		if params.ParentContext.IsSampled() {
			return trace.SamplingDecision{Sample: true}
		}

		// We have a parent that decided not to log; we don't log either.
		empty := trace.SpanContext{}
		if params.ParentContext != empty {
			return trace.SamplingDecision{}
		}

		sampler.mu.Lock()
		defer sampler.mu.Unlock()

		counter, ok := sampler.counters[params.Name]
		if !ok {
			counter = newWindowCounter()
			sampler.counters[params.Name] = counter
		}

		counter.incr(1)

		// We see an increase in the rate of traces, rate limit the endpoint.
		total := counter.total()
		if total > 20 {
			if counter.lastQuota.IsZero() {
				log.WithField("endpoint", params.Name).Info("Downgrade tracing to avoid sending too much data")
			}
			counter.lastQuota = time.Now()
		}

		// Standard case, no rate limiting measures are taken.
		if counter.lastQuota.IsZero() {
			return trace.SamplingDecision{Sample: true}
		}

		// If 48h make the endpoint quiet again we start tracing everything aggresively again.
		if time.Now().Sub(counter.lastQuota) > 48*time.Hour {
			log.WithField("endpoint", params.Name).Info("Upgrade tracing to always again")
			counter.lastQuota = time.Time{}
			return trace.SamplingDecision{Sample: true}
		}

		// We take 1 trace every 5-15 minutes randomly.
		if time.Now().After(counter.nextMeasurement) {
			counter.nextMeasurement = time.Now().Add(5*time.Minute + time.Duration(rand.Intn(10*60))*time.Second)
			return trace.SamplingDecision{Sample: true}
		}
		return trace.SamplingDecision{}
	}
}
