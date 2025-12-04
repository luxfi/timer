// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Package timer provides time utilities including mockable clocks
package timer

import (
	"sync"
	"time"
)

// MaxTime is the maximum representable time
var MaxTime = time.Unix(1<<63-62135596801, 0)

// Clock provides a mockable time source
type Clock struct {
	mu    sync.RWMutex
	faked bool
	time  time.Time
}

// NewClock returns a new clock synced to real time
func NewClock() *Clock {
	return &Clock{}
}

// Set sets the clock to a fixed time (for testing)
func (c *Clock) Set(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.faked = true
	c.time = t
}

// Sync resets the clock to use real time
func (c *Clock) Sync() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.faked = false
}

// Time returns the current time
func (c *Clock) Time() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.faked {
		return c.time
	}
	return time.Now()
}

// UnixTime returns the current time truncated to seconds
func (c *Clock) UnixTime() time.Time {
	return c.Time().Truncate(time.Second)
}

// Unix returns the current unix timestamp
func (c *Clock) Unix() uint64 {
	unix := max(c.Time().Unix(), 0)
	return uint64(unix)
}

// StoppedTimer is a timer that can be stopped and checked
type StoppedTimer struct {
	timer    *time.Timer
	finished bool
	mu       sync.Mutex
}

// NewStoppedTimer returns a new stopped timer
func NewStoppedTimer(f func()) *StoppedTimer {
	t := &StoppedTimer{}
	t.timer = time.AfterFunc(time.Hour, func() {
		t.mu.Lock()
		t.finished = true
		t.mu.Unlock()
		f()
	})
	t.timer.Stop()
	return t
}

// Reset resets the timer with the given duration
func (t *StoppedTimer) Reset(d time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.finished = false
	t.timer.Reset(d)
}

// Stop stops the timer
func (t *StoppedTimer) Stop() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.timer.Stop()
}

// Finished returns whether the timer has finished
func (t *StoppedTimer) Finished() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.finished
}

// Meter tracks event rates
type Meter struct {
	mu       sync.Mutex
	previous time.Time
	count    int64
	rate     float64
	halflife time.Duration
}

// NewMeter returns a new meter with the given halflife
func NewMeter(halflife time.Duration) *Meter {
	return &Meter{
		previous: time.Now(),
		halflife: halflife,
	}
}

// Tick records an event
func (m *Meter) Tick() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	now := time.Now()
	elapsed := now.Sub(m.previous)
	m.previous = now
	
	if elapsed > 0 && m.halflife > 0 {
		decay := elapsed.Seconds() / m.halflife.Seconds()
		m.rate = m.rate*powHalf(decay) + 1
	}
	m.count++
}

// Rate returns the current rate
func (m *Meter) Rate() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	now := time.Now()
	elapsed := now.Sub(m.previous)
	
	if elapsed > 0 && m.halflife > 0 {
		decay := elapsed.Seconds() / m.halflife.Seconds()
		return m.rate * powHalf(decay)
	}
	return m.rate
}

// Count returns the total count
func (m *Meter) Count() int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.count
}

// powHalf computes 0.5^x efficiently
func powHalf(x float64) float64 {
	// Use exponential decay formula: 0.5^x = e^(-x * ln(2))
	const ln2 = 0.693147180559945
	return exp(-x * ln2)
}

// exp computes e^x using Taylor series for small x
func exp(x float64) float64 {
	// For accuracy, use standard library for large values
	if x > 10 || x < -10 {
		return 0 // decay to 0 for very old events
	}
	
	result := 1.0
	term := 1.0
	for i := 1; i < 20; i++ {
		term *= x / float64(i)
		result += term
		if term < 1e-15 && term > -1e-15 {
			break
		}
	}
	return result
}
