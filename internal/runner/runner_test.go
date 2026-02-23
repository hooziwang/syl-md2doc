package runner

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"syl-md2doc/internal/job"
)

type fakeConverter struct {
	active    int32
	maxActive int32
}

func (f *fakeConverter) Convert(ctx context.Context, task job.Task) job.Result {
	cur := atomic.AddInt32(&f.active, 1)
	defer atomic.AddInt32(&f.active, -1)
	for {
		old := atomic.LoadInt32(&f.maxActive)
		if cur <= old || atomic.CompareAndSwapInt32(&f.maxActive, old, cur) {
			break
		}
	}

	time.Sleep(20 * time.Millisecond)
	if task.SourcePath == "bad" {
		return job.Result{Task: task, Error: fmt.Errorf("bad")}
	}
	return job.Result{Task: task, Warnings: []string{"w"}}
}

func TestRunnerContinueOnFailureAndCountSummary(t *testing.T) {
	c := &fakeConverter{}
	tasks := []job.Task{{SourcePath: "a"}, {SourcePath: "bad"}, {SourcePath: "c"}}
	s := Run(context.Background(), 2, tasks, c)
	require.Equal(t, 3, s.Total)
	require.Equal(t, 2, s.SuccessCount)
	require.Equal(t, 1, s.FailureCount)
	require.Equal(t, 2, s.WarningCount)
}

func TestRunnerRespectJobs(t *testing.T) {
	c := &fakeConverter{}
	tasks := make([]job.Task, 6)
	for i := 0; i < 6; i++ {
		tasks[i] = job.Task{SourcePath: "x"}
	}
	_ = Run(context.Background(), 2, tasks, c)
	require.LessOrEqual(t, int(c.maxActive), 2)
}
