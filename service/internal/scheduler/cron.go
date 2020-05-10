package scheduler

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"log"
	"time"
)

// Cron determines our sequence of jobs to execute
type Cron struct {
	Jobs []Job
}

func (c Cron) Halt(_ context.Context) error {
	log.Println("halting cron...")

	return nil
}

// Run continuously executes our sequence of Jobs
func (c Cron) Run(ctx context.Context) error {
	log.Println("running cron...")

	var runJob = func(ctx context.Context, j Job, ch chan Job) {
		result, err := j.Run(ctx)
		if err != nil {
			log.Println(errors.Wrap(err, fmt.Sprintf("ERROR job: [%s]", j.Name())))
		}
		log.Printf("job [%s]: %s", j.Name(), result)

		time.Sleep(time.Duration(j.IntervalInSeconds()) * time.Second)
		ch <- j
	}

	var jobC = make(chan Job)

	for _, j := range c.Jobs {
		go runJob(ctx, j, jobC)
	}

	for j := range jobC {
		go runJob(ctx, j, jobC)
	}

	return nil
}
