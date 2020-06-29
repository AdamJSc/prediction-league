package scheduler

import (
	"fmt"
	"github.com/pkg/errors"
)

// wrapJobStatus provides a helper function for wrapping the status of a job
func wrapJobStatus(j job, status string) string {
	return fmt.Sprintf("[%s] job status: %s", j.name(), status)
}

// wrapJobError provides a helper function for wrapping the details of an error encountered by a job
func wrapJobError(j job, err error) error {
	return errors.Wrapf(err, "[%s] error running job", j.name())
}
