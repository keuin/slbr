package common

/*
Copied from https://ixday.github.io/post/golang-cancel-copy/
*/

import (
	"context"
	"io"
)

// here is some syntax sugar inspired by the Tomas Senart's video,
// it allows me to inline the Reader interface
type readerFunc func(p []byte) (n int, err error)

func (rf readerFunc) Read(p []byte) (n int, err error) { return rf(p) }

// Copy slightly modified function signature:
// - context has been added in order to propagate cancellation
// - (undo by Keuin) I do not return the number of bytes written, has it is not useful in my use case
func Copy(ctx context.Context, out io.Writer, in io.Reader) (written int64, err error) {
	// Copy will call the Reader and Writer interface multiple time, in order
	// to copy by chunk (avoiding loading the whole file in memory).
	// I insert the ability to cancel before read time as it is the earliest
	// possible in the call process.
	written, err = io.Copy(out, readerFunc(func(p []byte) (int, error) {

		// golang non-blocking channel: https://gobyexample.com/non-blocking-channel-operations
		select {

		// if context has been canceled
		case <-ctx.Done():
			// stop process and propagate "context canceled" error
			return 0, ctx.Err()
		default:
			// otherwise just run default io.Reader implementation
			return in.Read(p)
		}
	}))

	return
}
