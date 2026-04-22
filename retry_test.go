package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

//
// ==========================
// Helpers for tests
// ==========================
//

func successFn() error {
	return nil
}

func failFn() error {
	return errors.New("fail")
}

func successWithValue(a int) (int, error) {
	return a * 2, nil
}

func failWithValue(a int) (int, error) {
	return 0, errors.New("fail")
}

//
// ==========================
// DoFunc tests
// ==========================
//

func TestDoFunc_Success(t *testing.T) {
	ctx := context.Background()

	err := DoFunc(ctx, 3, NewBackoff(10*time.Millisecond), successFn)
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestDoFunc_EventualSuccess(t *testing.T) {
	ctx := context.Background()

	attempts := 0
	fn := func() error {
		attempts++
		if attempts < 2 {
			return errors.New("fail")
		}
		return nil
	}

	err := DoFunc(ctx, 3, NewBackoff(5*time.Millisecond), fn)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestDoFunc_Failure(t *testing.T) {
	ctx := context.Background()

	err := DoFunc(ctx, 2, NewBackoff(1*time.Millisecond), failFn)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDoFunc_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	fn := func() error {
		cancel()
		return errors.New("fail")
	}

	err := DoFunc(ctx, 5, NewBackoff(10*time.Millisecond), fn)
	if err == nil {
		t.Fatal("expected context cancel error")
	}
}

//
// ==========================
// Do (reflection) tests
// ==========================
//

func TestDo_Success(t *testing.T) {
	ctx := context.Background()

	res, err := Do(ctx, 3, NewBackoff(10*time.Millisecond), successWithValue, 10)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if len(res) != 1 {
		t.Fatalf("expected 1 result, got %d", len(res))
	}

	if res[0].(int) != 20 {
		t.Fatalf("expected 20, got %v", res[0])
	}
}

func TestDo_Failure(t *testing.T) {
	ctx := context.Background()

	_, err := Do(ctx, 2, NewBackoff(1*time.Millisecond), failWithValue, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDo_InvalidFunction(t *testing.T) {
	ctx := context.Background()

	_, err := Do(ctx, 2, NewBackoff(1*time.Millisecond), 123)
	if err == nil {
		t.Fatal("expected error for invalid function")
	}
}

func TestDo_ArgumentMismatch(t *testing.T) {
	ctx := context.Background()

	fn := func(a int) (int, error) {
		return a, nil
	}

	_, err := Do(ctx, 2, NewBackoff(1*time.Millisecond), fn) // missing arg
	if err == nil {
		t.Fatal("expected argument mismatch error")
	}
}

func TestDo_LastReturnMustBeError(t *testing.T) {
	ctx := context.Background()

	fn := func() int {
		return 10
	}

	_, err := Do(ctx, 2, NewBackoff(1*time.Millisecond), fn)
	if err == nil {
		t.Fatal("expected error for missing error return")
	}
}

//
// ==========================
// Backoff tests
// ==========================
//

func TestBackoff_DurationBounds(t *testing.T) {
	b := NewBackoff(10 * time.Millisecond)
	b.Max = 50 * time.Millisecond

	d1 := b.Duration(1)
	d2 := b.Duration(10)

	if d1 <= 0 {
		t.Fatal("expected positive duration")
	}

	if d2 > b.Max {
		t.Fatal("expected capped max duration")
	}
}
