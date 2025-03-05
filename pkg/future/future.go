// pkg/future/future.go

package future

// Future[T] represents the result of an asynchronous operation.
type Future[T any] struct {
	resultChan chan T
	errChan    chan error
}

// NewFuture executes fn asynchronously and returns a Future[T].
func NewFuture[T any](fn func() (T, error)) Future[T] {
	f := Future[T]{
		resultChan: make(chan T, 1),
		errChan:    make(chan error, 1),
	}
	go func() {
		res, err := fn()
		if err != nil {
			f.errChan <- err
			return
		}
		f.resultChan <- res
	}()
	return f
}

// Await waits for and returns the result of the Future.
func (f Future[T]) Await() (T, error) {
	select {
	case res := <-f.resultChan:
		return res, nil
	case err := <-f.errChan:
		var zero T
		return zero, err
	}
}

// Bind applies fn to the result of f and returns a new Future[U].
func Bind[T any, U any](f Future[T], fn func(T) (U, error)) Future[U] {
	return NewFuture(func() (U, error) {
		res, err := f.Await()
		if err != nil {
			var zero U
			return zero, err
		}
		return fn(res)
	})
}

// Then is similar to Bind but for functions that don't return an error.
func Then[T any, U any](f Future[T], fn func(T) U) Future[U] {
	return Bind(f, func(res T) (U, error) {
		return fn(res), nil
	})
}
