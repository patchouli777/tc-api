package taskqueue

import (
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

type Scheduler struct {
	sched *asynq.Scheduler
}

func NewScheduler(sched *asynq.Scheduler) *Scheduler {
	return &Scheduler{sched: sched}
}

func (s *Scheduler) Schedule(timeout time.Duration,
	taskType string,
	payload []byte,
	taskId string) (string, error) {
	cronspec := fmt.Sprintf("@every %s", timeout.String())
	return s.sched.Register(cronspec, asynq.NewTask(taskType, payload), asynq.TaskID(taskId))
}

func (s *Scheduler) Run() error {
	return s.sched.Run()
}

func (s *Scheduler) Unregister(entryId string) error {
	return s.sched.Unregister(entryId)
}
