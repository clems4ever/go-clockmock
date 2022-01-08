# SmartClock

It is usually hard to test code that relies on real clocks because it hard to control time
and how much time is spent in functions or whether two events happening at the same time
appears in the same order at every run.

To solve that problem, smartclock allows to drive time from the tests in a deterministic
way so that there is no more randomness due to time in your tests. This greatly improves
the experience of writting tests because everything is deterministic, meaning reproducible
and easily debuggable.

## Get Started

Import smartclock into your project

```bash
go get github.com/clems4ever/go-smartclock
```

Wherever you need a clock, you should use the smartclock.Clock interface. That way, the
framework is able to smartly injects the smart mocked clock and allow you to drive it
from your tests.

Then you can inject a real clock in production code like so

```go
SetupTimerIn(&smartclock.RealClock{}, 20 * time.Minute, doSomething)
```

And in your tests, just use a mock. Obviously, the test will execute instaneously and have the
expected behavior when it comes to the timers being triggered.

```go
clockMock := smartclock.Mock(t, time.Date(2023, 4, 15, 11, 45, 0, 0, time.UTC))

SetupTimerIn(clockMock, 10 * time.Minute, doSomething)

// Move the clock forward 30 minutes, meaning that it will call doSomething after 10 minutes
// and keep moving forward until it reaches the target date.
// If clock.Now() is called in doSomething, it will return 2023-04-15 11:55:00.
clockMock.MoveForward(30 * time.Minutes)

// However, at this stage, i.e., after the clock has been moved forward,
// clockMock.Now() returns 2023-04-15 12:15:00 
```

For a concrete example, check the [examples/](./examples/) directory.

To leverage the maximum capacity of SmartClock, we advise to avoid the use of the After()
function altogether and prefer using AfterFunc() instead because it is way harder to control
the occurence of an event with channels than with deterministic function calls which is what
smartclock relies on through the use of a priority queue.

## How it works

The framework provides a clock interface `smartclock.Clock` that can be implemented by the real
`time.Time` clock in production code and by a mocked clock that can be manually driven in test
code.

This mock clock is essentially a clock holding a queue of timers supposed to trigger in the future.
When moving the clock forward, the clock checks whether some timers are supposed to trigger along
the way and it triggers them in order as if they were called at the expected time. This means that
if you call clock.Now() from within the time handler, it will return the time when the timer is
supposed to trigger.
If multiple timers are supposed to trigger until the target time, then they are sequentially
triggered at their respective time.
If multiple timers are a supposed to trigger at the same time, they are triggered sequentially in the
order they were started.

## License

This library is licensed under the MIT license.