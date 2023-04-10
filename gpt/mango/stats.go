package mango

import (
	"sync/atomic"
	"time"
)

type (
	Counter struct {
		atomic.Uint64
	}

	CacheLoadStats struct {
		Success Counter
		Failure Counter
		InNanos Counter
	}

	CacheStats struct {
		Hit, Miss Counter

		Eviction, EvictionExplicit, EvictionMaxSize Counter

		Load CacheLoadStats
	}
)

func NewCounter(value uint64) Counter {
	var v atomic.Uint64
	v.Store(value)
	return Counter{v}
}

func (c *Counter) Count(delta uint64) uint64 {
	return c.Uint64.Add(delta)
}

func (lcs *CacheLoadStats) AverageLoadPenalty() time.Duration {
	if total := lcs.TotalLoadCount(); total != 0 {
		nanos := lcs.InNanos.Count(0)
		return time.Duration(nanos / total)
	}
	return time.Duration(0)
}

func (lcs *CacheLoadStats) LoadExceptionRate() float64 {
	if total := float64(lcs.TotalLoadCount()); total != 0 {
		count := float64(lcs.Failure.Count(0))
		return count / total
	}
	return 0
}

func (lcs *CacheLoadStats) TotalLoadCount() uint64 {
	bad := lcs.Failure.Count(0)
	return lcs.Success.Count(0) + bad
}

func (cs *CacheStats) HitRate() float64 {
	if total := float64(cs.Total()); total != 0 {
		count := float64(cs.Hit.Count(0))
		return count / total
	}
	return 0
}

func (cs *CacheStats) MissRate() float64 {
	if total := float64(cs.Total()); total != 0 {
		count := float64(cs.Miss.Count(0))
		return count / total
	}
	return 0
}

func (cs *CacheStats) Total() uint64 {
	bad := cs.Miss.Count(0)
	return cs.Hit.Count(0) + bad
}
