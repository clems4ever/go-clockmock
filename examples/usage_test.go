package usage_test

import (
	"testing"
	"time"

	"github.com/clems4ever/go-smartclock"
	"github.com/stretchr/testify/require"
)

func TestDrivableTimer(t *testing.T) {
	var called bool
	clockMock := smartclock.Mock(t, smartclock.ExampleBaseTime)

	clockMock.AfterFunc(20*time.Minute, func() {
		called = true
	})

	require.False(t, called)

	// Moving the clock forward 19 minutes
	clockMock.MoveForward(19 * time.Minute)
	require.False(t, called)

	// And forward 59 seconds.
	clockMock.MoveForward(59 * time.Second)
	require.False(t, called)

	// And then it should trigger in 1 second.
	clockMock.MoveForward(time.Second)
	require.True(t, called)
}
