package timebank

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidParameters = errors.New("timebank: invalid parameters")
	ErrInvalidDeadline   = errors.New("timebank: invalid deadline")
)

const (
	DefaultTimeout time.Duration = 15 * time.Second
)

type TimeBank struct {
	isRunning bool
	timer     *time.Timer
	due       time.Time
	callback  func(bool)
	cancel    func()
}

func NewTimeBank() *TimeBank {

	// Initializing timer
	timer := time.NewTimer(DefaultTimeout)
	timer.Stop()

	tb := &TimeBank{
		isRunning: false,
		timer:     timer,
		callback:  func(bool) {},
		cancel:    func() {},
	}

	return tb
}

func (tb *TimeBank) Cancel() {
	tb.isRunning = false
	tb.timer.Stop()
	tb.cancel()
}

func (tb *TimeBank) NewTask(duration time.Duration, fn func(isCancelled bool)) error {

	if fn == nil {
		return ErrInvalidParameters
	}

	tb.Cancel()

	// Trigger immediately
	if duration == time.Second*0 && fn != nil {
		tb.callback = fn
		fn(false)
		return nil
	}

	// Initializing context
	ctx, cancel := context.WithCancel(context.Background())
	tb.cancel = cancel
	tb.isRunning = true
	tb.callback = fn
	tb.due = time.Now().Add(duration)
	tb.timer.Reset(duration)

	go func(ctx context.Context) {
		select {
		case <-tb.timer.C:
			tb.callback(false)
		case <-ctx.Done():
			tb.callback(true)
		}
	}(ctx)

	return nil
}

func (tb *TimeBank) Extend(duration time.Duration) bool {

	// Time bank is not running
	if !tb.isRunning || tb.due.Before(time.Now()) {
		return false
	}

	// total = remain + extend
	total := tb.due.Sub(time.Now()) + duration

	// Update timer
	tb.due = tb.due.Add(duration)
	tb.timer.Reset(total)

	return true
}

func (tb *TimeBank) NewTaskWithDeadline(deadline time.Time, fn func(isCancelled bool)) error {

	now := time.Now()

	if deadline.Before(now) {
		return ErrInvalidDeadline
	}

	// Calculate duration
	duration := deadline.Sub(now)

	return tb.NewTask(duration, fn)
}
