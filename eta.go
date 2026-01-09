// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package timer

import (
	"encoding/binary"
	"time"
)

// ProgressFromHash returns the progress out of MaxUint64 assuming [b] is a key
// in a uniformly distributed sequence that is being iterated lexicographically.
func ProgressFromHash(b []byte) uint64 {
	// binary.BigEndian.Uint64 will panic if the input length is less than 8, so
	// pad 0s as needed.
	var progress [8]byte
	copy(progress[:], b)
	return binary.BigEndian.Uint64(progress[:])
}

// EstimateETA estimates the remaining time to complete a task given
// the startTime and its current progress.
func EstimateETA(startTime time.Time, progress, end uint64) time.Duration {
	if progress == 0 || end == 0 {
		return 0
	}
	timeSpent := time.Since(startTime)

	percentExecuted := float64(progress) / float64(end)
	estimatedTotalDuration := time.Duration(float64(timeSpent) / percentExecuted)
	eta := estimatedTotalDuration - timeSpent
	return eta.Round(time.Second)
}

// EtaTracker provides exponentially weighted moving average ETA estimates
type EtaTracker struct {
	startTime  time.Time
	lastUpdate time.Time
	progress   uint64
	total      uint64
	ewmaRate   float64 // exponentially weighted moving average of rate
	alpha      float64 // smoothing factor for EWMA
}

// NewEtaTracker creates a new ETA tracker with the given total and smoothing factor.
// Alpha should be between 0 and 1; higher values give more weight to recent observations.
func NewEtaTracker(total uint64, alpha float64) *EtaTracker {
	if alpha <= 0 || alpha > 1 {
		alpha = 0.3 // default smoothing factor
	}
	return &EtaTracker{
		startTime:  time.Now(),
		lastUpdate: time.Now(),
		total:      total,
		alpha:      alpha,
	}
}

// Update records new progress and returns the estimated time remaining.
func (e *EtaTracker) Update(progress uint64) time.Duration {
	now := time.Now()
	elapsed := now.Sub(e.lastUpdate)

	if elapsed > 0 && progress > e.progress {
		// Calculate instantaneous rate
		delta := progress - e.progress
		rate := float64(delta) / elapsed.Seconds()

		// Update EWMA rate
		if e.ewmaRate == 0 {
			e.ewmaRate = rate
		} else {
			e.ewmaRate = e.alpha*rate + (1-e.alpha)*e.ewmaRate
		}
	}

	e.progress = progress
	e.lastUpdate = now

	remaining := e.total - progress
	if e.ewmaRate > 0 {
		eta := time.Duration(float64(remaining) / e.ewmaRate * float64(time.Second))
		return eta.Round(time.Second)
	}

	// Fall back to simple estimate
	return EstimateETA(e.startTime, progress, e.total)
}

// Progress returns the current progress.
func (e *EtaTracker) Progress() uint64 {
	return e.progress
}

// Total returns the total amount of work.
func (e *EtaTracker) Total() uint64 {
	return e.total
}

// Rate returns the current estimated rate per second.
func (e *EtaTracker) Rate() float64 {
	return e.ewmaRate
}
