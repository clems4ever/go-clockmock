package smartclock

import (
	"container/heap"
	"sync"
	"testing"
	"time"
)

var ExampleBaseTime = time.Date(2023, 5, 23, 10, 45, 30, 0, time.UTC)

type MockTimer struct {
	// this ID is used to execute timers with same date in the order they were initially inserted in.
	ID   int
	Date time.Time
	Func func()

	mu sync.Mutex

	resetFn func(d time.Duration)
	stopFn  func()
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*MockTimer

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	if pq[i].Date.Equal(pq[j].Date) {
		return pq[i].ID < pq[j].ID
	}
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].Date.Before(pq[j].Date)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x any) {
	item := x.(*MockTimer)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}

func (mt *MockTimer) Stop() bool {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	mt.stopFn()
	return true
}

func (mt *MockTimer) Reset(d time.Duration) bool {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	mt.resetFn(d)
	return true
}

type MockClock struct {
	t *testing.T

	baseTimeMu sync.RWMutex
	baseTime   time.Time

	timersCounterMu sync.Mutex
	timersCounter   int

	activeTimersMu    sync.RWMutex
	activeTimersQueue PriorityQueue

	inactiveTimersMu sync.RWMutex
	inactiveTimers   map[*MockTimer]struct{}

	updatedTimerUpdatesMu sync.RWMutex
	updatedTimerUpdates   map[*MockTimer]MockTimerUpdate
}

type MockTimerUpdateType int

const (
	ResetMockTimerUpdateType MockTimerUpdateType = iota
	StopMockTimerUpdateType  MockTimerUpdateType = iota
)

type MockTimerUpdate struct {
	Type MockTimerUpdateType
	Date time.Time
}

func Mock(t *testing.T, baseTime time.Time) *MockClock {
	mc := &MockClock{
		t:                   t,
		baseTime:            baseTime,
		inactiveTimers:      make(map[*MockTimer]struct{}),
		updatedTimerUpdates: make(map[*MockTimer]MockTimerUpdate),
	}
	heap.Init(&mc.activeTimersQueue)
	return mc
}

func (mc *MockClock) MoveTo(targetTime time.Time) {
	for {
		mc.activeTimersMu.Lock()
		if mc.activeTimersQueue.Len() == 0 {
			mc.activeTimersMu.Unlock()
			break
		}

		nextTimer := heap.Pop(&mc.activeTimersQueue).(*MockTimer)
		mc.activeTimersMu.Unlock()

		if nextTimer.Date.After(targetTime) {
			// if the timer is beyond the targetTime it means that there is no timer to execute anymore.
			mc.activeTimersMu.Lock()
			heap.Push(&mc.activeTimersQueue, nextTimer)
			mc.activeTimersMu.Unlock()
			break
		}

		mc.baseTimeMu.Lock()
		mc.baseTime = nextTimer.Date
		mc.baseTimeMu.Unlock()

		mc.updatedTimerUpdatesMu.Lock()
		mc.updatedTimerUpdates[nextTimer] = MockTimerUpdate{
			Type: StopMockTimerUpdateType,
		}
		mc.updatedTimerUpdatesMu.Unlock()

		nextTimer.Func()
		mc.processUpdatedTimers()
	}

	mc.baseTimeMu.Lock()
	mc.baseTime = targetTime
	mc.baseTimeMu.Unlock()
}

func (mc *MockClock) processUpdatedTimers() {
	mc.updatedTimerUpdatesMu.Lock()
	defer mc.updatedTimerUpdatesMu.Unlock()

	for timer, update := range mc.updatedTimerUpdates {
		switch update.Type {
		case ResetMockTimerUpdateType:
			timer.Date = update.Date
			if mc.isTimerActive(timer) {
				mc.activeTimersMu.Lock()
				heap.Init(&mc.activeTimersQueue)
				mc.activeTimersMu.Unlock()
			} else {
				mc.activeTimersMu.Lock()
				heap.Push(&mc.activeTimersQueue, timer)
				mc.activeTimersMu.Unlock()
			}
		case StopMockTimerUpdateType:
			if mc.isTimerActive(timer) {
				mc.deleteFromActiveTimers(timer)
				mc.pushToInactiveTimers(timer)
			}
		}
	}
}

// MoveForward moves the clock forward by a duration D.
func (mc *MockClock) MoveForward(d time.Duration) {
	mc.MoveTo(mc.Now().Add(d))
}

// Now returns the current date.
func (mc *MockClock) Now() time.Time {
	mc.baseTimeMu.RLock()
	defer mc.baseTimeMu.RUnlock()
	return mc.baseTime
}

func (mc *MockClock) isTimerActive(timer *MockTimer) bool {
	mc.activeTimersMu.RLock()
	defer mc.activeTimersMu.RUnlock()

	for _, qtimer := range mc.activeTimersQueue {
		if qtimer == timer {
			return true
		}
	}
	return false
}

func (mc *MockClock) pushToInactiveTimers(timer *MockTimer) {
	mc.inactiveTimersMu.Lock()
	defer mc.inactiveTimersMu.Unlock()
	mc.inactiveTimers[timer] = struct{}{}
}

func (mc *MockClock) deleteFromActiveTimers(timer *MockTimer) {
	mc.activeTimersMu.Lock()
	defer mc.activeTimersMu.Unlock()

	var newActiveTimers PriorityQueue
	for _, atimer := range mc.activeTimersQueue {
		if atimer == timer {
			continue
		}
		newActiveTimers = append(newActiveTimers, atimer)
	}
	heap.Init(&newActiveTimers)
	mc.activeTimersQueue = newActiveTimers
}

func (mc *MockClock) AfterFunc(duration time.Duration, fn func()) Timer {
	timer := &MockTimer{
		Date: mc.baseTime.Add(duration),
		Func: fn,
	}

	mc.timersCounterMu.Lock()
	timer.ID = mc.timersCounter
	mc.timersCounter += 1
	mc.timersCounterMu.Unlock()

	timer.resetFn = func(d time.Duration) {
		mc.updatedTimerUpdatesMu.Lock()
		mc.updatedTimerUpdates[timer] = MockTimerUpdate{
			Type: ResetMockTimerUpdateType,
			Date: mc.baseTime.Add(d),
		}
		mc.updatedTimerUpdatesMu.Unlock()
	}
	timer.stopFn = func() {
		mc.updatedTimerUpdatesMu.Lock()
		mc.updatedTimerUpdates[timer] = MockTimerUpdate{
			Type: StopMockTimerUpdateType,
		}
		mc.updatedTimerUpdatesMu.Unlock()
	}

	mc.activeTimersMu.Lock()
	mc.activeTimersQueue.Push(timer)
	mc.activeTimersMu.Unlock()
	return timer
}

func (mc *MockClock) After(d time.Duration) <-chan time.Time {
	timeC := make(chan time.Time)
	timer := &MockTimer{
		Date: mc.baseTime.Add(d),
		Func: func() {
			timeC <- mc.baseTime
		},
	}
	mc.activeTimersMu.Lock()
	mc.activeTimersQueue.Push(timer)
	mc.activeTimersMu.Unlock()
	return timeC
}
