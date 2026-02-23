package runner

import (
	"context"
	"sync"

	"syl-md2doc/internal/convert"
	"syl-md2doc/internal/job"
)

func Run(ctx context.Context, jobs int, tasks []job.Task, c convert.Converter) Summary {
	if jobs < 1 {
		jobs = 1
	}
	if len(tasks) == 0 {
		return Summary{}
	}

	type indexedResult struct {
		idx int
		res job.Result
	}

	workCh := make(chan int)
	resultCh := make(chan indexedResult, len(tasks))
	wg := sync.WaitGroup{}

	for i := 0; i < jobs; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range workCh {
				res := c.Convert(ctx, tasks[idx])
				resultCh <- indexedResult{idx: idx, res: res}
			}
		}()
	}

	go func() {
		for i := range tasks {
			workCh <- i
		}
		close(workCh)
		wg.Wait()
		close(resultCh)
	}()

	results := make([]job.Result, len(tasks))
	for item := range resultCh {
		results[item.idx] = item.res
	}

	summary := Summary{Total: len(tasks), Results: results}
	for _, r := range results {
		summary.WarningCount += len(r.Warnings)
		if r.Error != nil {
			summary.FailureCount++
			continue
		}
		summary.SuccessCount++
	}
	return summary
}
