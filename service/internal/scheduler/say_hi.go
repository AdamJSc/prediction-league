package scheduler

import (
	"context"
	"fmt"
)

type SayHiJob struct {
	Job
}

func (s SayHiJob) Name() string {
	return "say_hi_job"
}

func (s SayHiJob) IntervalInSeconds() int {
	return 5
}

func (s SayHiJob) Run(_ context.Context) (string, error) {
	return fmt.Sprintf("heeeeeeyy!! repeating in %d seconds....\n", s.IntervalInSeconds()), nil
}
