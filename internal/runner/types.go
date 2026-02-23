package runner

import "syl-md2doc/internal/job"

type Summary struct {
	Total        int
	SuccessCount int
	FailureCount int
	WarningCount int
	Results      []job.Result
}
