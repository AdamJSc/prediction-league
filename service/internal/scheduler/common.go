package scheduler

import (
	"github.com/pkg/errors"
)

// wrapJobError provides a helper function for wrapping the details of an error encountered by a job
func wrapJobError(jobName string, err error) error {
	return errors.Wrapf(err, "[%s] error running job", jobName)
}
