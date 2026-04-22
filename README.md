# retry

A simple, production-ready retry library for Go with exponential backoff, jitter, and context support.

## Features

- Simple API for retrying operations
- Context-aware cancellation & timeout support
- Exponential backoff with jitter
- Safe, production-tested design
- No global state
- Lightweight and dependency-free

## Installation

```bash
go get github.com/thedevsaddam/retry/v2
```

## Usage

### Retry a function returning error

```go
err := retry.DoFunc(ctx, 3, retry.NewBackoff(100*time.Millisecond), func() error {
    return doSomething()
})

if err != nil {
    log.Fatal(err)
}
```

### Retry with a generic function

```go
res, err := retry.Do(ctx, 3, retry.NewBackoff(100*time.Millisecond), myFunc, arg1, arg2)
if err != nil {
    log.Fatal(err)
}

fmt.Println(res)
```

## Backoff configuration

```go
b := retry.NewBackoff(100 * time.Millisecond)
b.Max = 5 * time.Second
b.Jitter = 0.3
b.Factor = 2
```

## Behavior

- Retries stop immediately on success
- Retries stop on context cancellation
- Exponential backoff increases delay per attempt
- Jitter prevents thundering herd problems
- Max backoff prevents unbounded wait time

## Example

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := retry.DoFunc(ctx, 5, retry.NewBackoff(200*time.Millisecond), func() error {
    // API call / DB call / network request
    return errors.New("temporary failure")
})
```

## Design Goals

- Simple API surface
- Safe for production systems
- Predictable retry behavior
- Easy to embed in microservices

## License

MIT
