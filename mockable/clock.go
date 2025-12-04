// Copyright (C) 2019-2025, Lux Industries, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Package mockable provides a mockable clock for testing.
package mockable

import "time"

// MaxTime is the maximum representable time.
var MaxTime = time.Unix(1<<63-62135596801, 0)

// Clock acts as a thin wrapper around global time that allows for easy testing.
type Clock struct {
	faked bool
	time  time.Time
}

// Set the time on the clock.
func (c *Clock) Set(time time.Time) { c.faked = true; c.time = time }

// Sync this clock with global time.
func (c *Clock) Sync() { c.faked = false }

// Time returns the time on this clock.
func (c *Clock) Time() time.Time {
	if c.faked {
		return c.time
	}
	return time.Now()
}

// UnixTime returns the unix time on this clock truncated to seconds.
func (c *Clock) UnixTime() time.Time {
	return c.Time().Truncate(time.Second)
}

// Unix returns the unix timestamp on this clock.
func (c *Clock) Unix() uint64 {
	unix := max(c.Time().Unix(), 0)
	return uint64(unix)
}
