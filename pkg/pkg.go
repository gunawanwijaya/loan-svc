package pkg

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"sync"
)

type noCopy struct{}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Nonce(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}

func Must1[T any](t1 T, err error) (_ T) { Must(err); return t1 }

func Must2[T1, T2 any](t1 T1, t2 T2, err error) (_ T1, _ T2) { Must(err); return t1, t2 }

func OrElse[T any](c bool, v1 T, v2 T) T {
	if c {
		return v1
	}
	return v2
}

func Sync(mu sync.Locker, fn func()) { mu.Lock(); fn(); mu.Unlock() }

func AsValidator[T Validator[T]](t T) T { return t }

func Validate[T any](ctx context.Context, t Validator[T]) (T, error) { return t.Validate(ctx) }

type Validator[T any] interface {
	Validate(ctx context.Context) (T, error)
}

func Callstack() *callstack {
	return &callstack{
		stack: nil,
		once:  &sync.Once{},
		errCh: make(chan error, 1),
	}
}

type callstack struct {
	stack []func()
	once  *sync.Once
	errCh chan error
}

func (x *callstack) Wait() error { return <-x.errCh }

func (x *callstack) Call(ctx context.Context, cancel context.CancelFunc) {
	x.once.Do(func() {
		defer cancel()
		go func() { <-ctx.Done(); x.errCh <- ctx.Err() }()
		for i := len(x.stack) - 1; i >= 0; i-- {
			if x.stack[i] != nil {
				x.stack[i]()
				// <-time.After(5 * time.Second)
			}
		}
	})
}

func (x *callstack) Register(callback func()) {
	if callback != nil {
		x.stack = append(x.stack, callback)
	}
}

func BtoA(p []byte) string { return base64.StdEncoding.EncodeToString(p) }
func AtoB(s string) []byte { p, _ := base64.StdEncoding.DecodeString(s); return p }
func Ptr[T any](v T) *T    { return &v }
