package taskqueue

import (
	"context"

	"github.com/hibiken/asynq"
)

func TaskHandler(f func(context.Context, []byte) error) func(ctx context.Context, task *asynq.Task) error {
	return func(ctx context.Context, task *asynq.Task) error {
		return f(ctx, task.Payload())
	}
}
