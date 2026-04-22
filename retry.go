// Package retry provides a production-ready retry utility with
// exponential backoff, jitter, context support, and optional reflection.
package retry

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"reflect"
	"time"
)

var (
	ErrInvalidAttempt = errors.New("retry: attempt must be > 0")
	ErrFnNotFunc      = errors.New("retry: fn is not a function")
	ErrNoReturn       = errors.New("retry: function must return at least one value")
	ErrLastMustBeErr  = errors.New("retry: last return value must be error")
)

// Backoff controls retry delay behavior.
type Backoff struct {
	Base   time.Duration
	Max    time.Duration
	Jitter float64 // 0.0 - 1.0
	Factor float64

	rnd *rand.Rand
}

// NewBackoff creates a default exponential backoff.
func NewBackoff(base time.Duration) *Backoff {
	return &Backoff{
		Base:   base,
		Max:    30 * time.Second,
		Jitter: 0.3,
		Factor: 2,
		rnd:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Duration calculates backoff duration for a given attempt.
func (b *Backoff) Duration(attempt int) time.Duration {
	if b.Factor <= 0 {
		b.Factor = 2
	}

	// exponential base
	d := float64(b.Base) * pow(b.Factor, float64(attempt))

	// hard cap BEFORE jitter
	if d > float64(b.Max) {
		d = float64(b.Max)
	}

	// jitter
	j := d * b.Jitter
	d = d - j + (b.rnd.Float64() * 2 * j)

	// 🚨 final cap AFTER jitter (IMPORTANT)
	if d > float64(b.Max) {
		d = float64(b.Max)
	}

	return time.Duration(d)
}

// simple power (avoids math.Pow overhead for small values)
func pow(a, b float64) float64 {
	result := 1.0
	for i := 0; i < int(b); i++ {
		result *= a
	}
	return result
}

//
// =============================
// Typed retry (recommended)
// =============================
//

// DoFunc retries a function returning only error.
func DoFunc(
	ctx context.Context,
	attempts int,
	backoff *Backoff,
	fn func() error,
) error {

	if attempts <= 0 {
		return ErrInvalidAttempt
	}

	if backoff == nil {
		backoff = NewBackoff(100 * time.Millisecond)
	}

	var err error

	for i := 0; i < attempts; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err = fn(); err == nil {
			return nil
		}

		if i < attempts-1 {
			if sleep(ctx, backoff.Duration(i)) != nil {
				return ctx.Err()
			}
		}
	}

	return err
}

//
// =============================
// Reflect-based retry (flexible)
// =============================
//

// Do retries any function (last return value MUST be error).
func Do(
	ctx context.Context,
	attempts int,
	backoff *Backoff,
	fn interface{},
	args ...interface{},
) ([]interface{}, error) {

	if attempts <= 0 {
		return nil, ErrInvalidAttempt
	}

	vfn := reflect.ValueOf(fn)

	if vfn.Kind() != reflect.Func {
		return nil, ErrFnNotFunc
	}

	ft := vfn.Type()

	if ft.NumOut() == 0 {
		return nil, ErrNoReturn
	}

	if !ft.IsVariadic() && ft.NumIn() != len(args) {
		return nil, errors.New("retry: argument mismatch")
	}

	if backoff == nil {
		backoff = NewBackoff(100 * time.Millisecond)
	}

	in := make([]reflect.Value, len(args))
	for i, a := range args {
		in[i] = reflect.ValueOf(a)
	}

	var lastErr error

	for i := 0; i < attempts; i++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		out := vfn.Call(in)

		result := make([]interface{}, len(out))
		for i, o := range out {
			result[i] = o.Interface()
		}

		last := result[len(result)-1]

		err, ok := last.(error)
		if !ok && last != nil {
			return nil, ErrLastMustBeErr
		}

		if err == nil {
			return result[:len(result)-1], nil
		}

		lastErr = err

		if i < attempts-1 {
			if sleep(ctx, backoff.Duration(i)) != nil {
				return nil, ctx.Err()
			}
		}
	}

	return nil, lastErr
}

//
// =============================
// Helpers
// =============================
//

func sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// Optional compatibility (legacy style)
func init() {
	// no global rand seeding anymore (intentionally removed)
	fmt.Println("retry package initialized")
}
