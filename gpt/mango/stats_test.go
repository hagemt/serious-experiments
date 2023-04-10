package mango

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestZeros(t *testing.T) {
	none := &CacheStats{}
	assert.Equal(t, float64(0), none.HitRate())
	assert.Equal(t, float64(0), none.MissRate())
	assert.Equal(t, uint64(0), none.Total())
	assert.Equal(t, uint64(0), none.Load.TotalLoadCount())
	assert.Equal(t, float64(0), none.Load.LoadExceptionRate())
	assert.Equal(t, time.Duration(0), none.Load.AverageLoadPenalty())

	some := &CacheStats{
		Miss: NewCounter(1),
		Hit:  NewCounter(1),
	}
	assert.Equal(t, float64(0.5), some.HitRate())
	assert.Equal(t, float64(0.5), some.MissRate())
	assert.Equal(t, uint64(2), some.Total())

	more := &CacheLoadStats{
		Success: NewCounter(2),
		Failure: NewCounter(2),
		InNanos: NewCounter(uint64(time.Second)),
	}
	total := more.TotalLoadCount()
	half := more.LoadExceptionRate()
	nanos := more.AverageLoadPenalty()
	assert.Equal(t, time.Second/4, nanos)
	assert.Equal(t, float64(0.5), half)
	assert.Equal(t, uint64(4), total)
}
