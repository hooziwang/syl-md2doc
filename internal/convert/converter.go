package convert

import (
	"context"

	"syl-md2doc/internal/job"
)

type Converter interface {
	Convert(ctx context.Context, task job.Task) job.Result
}
