package scheduler

import (
	"fmt"
	"github.com/pkg/errors"
)

// wrapJobStatus provides a helper function for wrapping the status of a job
func wrapJobStatus(jobName string, status string) string {
	return fmt.Sprintf("[%s] job status: %s", jobName, status)
}

// wrapJobError provides a helper function for wrapping the details of an error encountered by a job
func wrapJobError(jobName string, err error) error {
	return errors.Wrapf(err, "[%s] error running job", jobName)
}
