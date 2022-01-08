package smartclock_test

import (
	"testing"
	"time"

	"github.com/clems4ever/go-smartclock"
	"github.com/stretchr/testify/require"
)

func TestClockShouldTriggerAfterFunc(t *testing.T) {
	clock := smartclock.Mock(t, smartclock.ExampleBaseTime)
	var called bool = false

	clock.AfterFunc(2*time.Minute, func() {
		called = true
	})

	clock.MoveForward(time.Minute)

	require.False(t, called)

	clock.MoveForward(time.Minute)

	require.True(t, called)
}

func TestClockShouldTriggerAfterFuncOfMultipleTimers(t *testing.T) {
	clock := smartclock.Mock(t, smartclock.ExampleBaseTime)
	var called1 bool = false
	var called2 bool = false

	clock.AfterFunc(2*time.Minute, func() {
		called1 = true
	})
	clock.AfterFunc(4*time.Minute, func() {
		called2 = true
	})

	clock.MoveForward(time.Minute)

	require.False(t, called1)
	require.False(t, called2)

	clock.MoveForward(time.Minute)

	require.True(t, called1)
	require.False(t, called2)

	clock.MoveForward(time.Minute)

	require.True(t, called1)
	require.False(t, called2)

	clock.MoveForward(time.Minute)

	require.True(t, called1)
	require.True(t, called2)
}

type Wrapper struct {
	timer smartclock.Timer
	Count int

	StopAfter int
	// finishes counting after this many iteration
	FinishAfter int
}

func (w *Wrapper) Start(clock smartclock.Clock) {
	w.timer = clock.AfterFunc(time.Minute, w.count)
}

func (w *Wrapper) count() {
	w.Count += 1

	if w.StopAfter != 0 && w.StopAfter <= w.Count {
		w.timer.Stop()
		return
	}

	if w.FinishAfter == 0 || w.FinishAfter > w.Count {
		w.timer.Reset(time.Minute)
	}
}

func TestClockShouldTriggerAfterFuncSubsequently(t *testing.T) {
	clock := smartclock.Mock(t, smartclock.ExampleBaseTime)
	wrapper := &Wrapper{}

	wrapper.Start(clock)

	clock.MoveForward(time.Minute)

	require.Equal(t, 1, wrapper.Count)

	clock.MoveForward(time.Minute)

	require.Equal(t, 2, wrapper.Count)
}

func TestClockShouldTriggerAfterFuncSubsequentlyAndFinish(t *testing.T) {
	clock := smartclock.Mock(t, smartclock.ExampleBaseTime)
	wrapper := &Wrapper{FinishAfter: 3}

	wrapper.Start(clock)

	clock.MoveForward(time.Minute)
	require.Equal(t, 1, wrapper.Count)

	clock.MoveForward(time.Minute)
	require.Equal(t, 2, wrapper.Count)

	clock.MoveForward(time.Minute)
	require.Equal(t, 3, wrapper.Count)

	clock.MoveForward(time.Minute)
	require.Equal(t, 3, wrapper.Count)
}

func TestClockShouldTriggerAfterFuncSubsequentlyAndStop(t *testing.T) {
	clock := smartclock.Mock(t, smartclock.ExampleBaseTime)
	wrapper := &Wrapper{StopAfter: 3}

	wrapper.Start(clock)

	clock.MoveForward(time.Minute)
	require.Equal(t, 1, wrapper.Count)

	clock.MoveForward(time.Minute)
	require.Equal(t, 2, wrapper.Count)

	clock.MoveForward(time.Minute)
	require.Equal(t, 3, wrapper.Count)

	clock.MoveForward(time.Minute)
	require.Equal(t, 3, wrapper.Count)
}

func TestClockShouldResetOneActiveTimerFromAnotherOne(t *testing.T) {
	clock := smartclock.Mock(t, smartclock.ExampleBaseTime)
	var called bool = false

	t1 := clock.AfterFunc(3*time.Minute, func() {
		called = true
	})

	clock.AfterFunc(time.Minute, func() {
		t1.Reset(3 * time.Minute)
	})

	require.False(t, called)
	clock.MoveForward(time.Minute)
	require.False(t, called)

	clock.MoveForward(time.Minute)
	require.False(t, called)

	clock.MoveForward(2 * time.Minute)
	require.True(t, called)
}

func TestClockShouldResetOneInactiveTimerFromAnotherOne(t *testing.T) {
	clock := smartclock.Mock(t, smartclock.ExampleBaseTime)
	var count int

	t1 := clock.AfterFunc(time.Minute, func() {
		count += 1
	})

	clock.AfterFunc(2*time.Minute, func() {
		t1.Reset(3 * time.Minute)
	})

	require.Equal(t, 0, count)
	clock.MoveForward(time.Minute)
	require.Equal(t, 1, count)

	clock.MoveForward(2 * time.Minute)
	require.Equal(t, 1, count)

	clock.MoveForward(3 * time.Minute)
	require.Equal(t, 2, count)
}
