package smartclock

import "time"

var Year = 365 * 24 * time.Hour

type Timer interface {
	Stop() bool
	Reset(d time.Duration) bool
}

//go:generate mockgen -destination=utils_mocks/clock_mock.go -package=utils_mocks github.com/AviatrixDev/scand/internal/utils Clock
type Clock interface {
	Now() time.Time
	AfterFunc(duration time.Duration, fn func()) Timer
	After(d time.Duration) <-chan time.Time
}

type RealClock struct{}

func (rc *RealClock) Now() time.Time {
	return time.Now()
}

func (c *RealClock) AfterFunc(duration time.Duration, fn func()) Timer {
	return &TimerWrapper{
		Timer: time.AfterFunc(duration, fn),
	}
}

func (rc *RealClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

type TimerWrapper struct {
	*time.Timer
}
