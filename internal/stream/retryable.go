package stream

import (
	"context"
	"github.com/cenkalti/backoff/v4"
)

// Retryable will initiate the stream listener. On error,
// such as when the server is unreachable, or another
// unrecoverable error occurs, this function will log it,
// initiate a backoff, and retry connection after some seconds
func Retryable(stream *Stream, ctx context.Context, notify backoff.Notify) error {
	return backoff.RetryNotify(func() error {
		return stream.Listen(ctx)
	}, backoff.NewExponentialBackOff(), notify)
}
