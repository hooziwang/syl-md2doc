package app

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"syl-md2doc/internal/convert"
	"syl-md2doc/internal/input"
	"syl-md2doc/internal/plan"
	"syl-md2doc/internal/runner"
)

func Run(opts Options) (Result, error) {
	if len(opts.Inputs) == 0 {
		return Result{}, fmt.Errorf("至少提供一个输入")
	}

	cwd := strings.TrimSpace(opts.CWD)
	if cwd == "" {
		wd, err := os.Getwd()
		if err != nil {
			return Result{}, fmt.Errorf("读取当前目录失败：%w", err)
		}
		cwd = wd
	}

	jobs := opts.Jobs
	if jobs <= 0 {
		jobs = runtime.NumCPU()
	}
	if jobs < 1 {
		jobs = 1
	}

	conv := opts.Converter
	if conv == nil {
		if err := convert.EnsurePandocAvailable(opts.PandocPath); err != nil {
			return Result{}, err
		}
		conv = convert.NewPandocConverter(opts.PandocPath, opts.ReferenceDocx, opts.Verbose)
	}

	sources, discoverWarns, discoverFails, err := input.Discover(opts.Inputs, cwd)
	if err != nil {
		return Result{}, err
	}

	tasks, planWarns, err := plan.BuildTargets(sources, plan.Options{OutputArg: opts.OutputArg, CWD: cwd})
	if err != nil {
		return Result{}, err
	}

	summary := runner.Run(context.Background(), jobs, tasks, conv)

	result := Result{
		SuccessCount: summary.SuccessCount,
		Warnings:     make([]string, 0),
		Failures:     make([]Failure, 0),
	}
	result.Warnings = append(result.Warnings, discoverWarns...)
	result.Warnings = append(result.Warnings, planWarns...)

	for _, f := range discoverFails {
		result.Failures = append(result.Failures, Failure{Source: f.Input, Reason: f.Reason})
	}
	for _, item := range summary.Results {
		result.Warnings = append(result.Warnings, item.Warnings...)
		if item.Error != nil {
			result.Failures = append(result.Failures, Failure{Source: item.Task.SourcePath, Reason: item.Error.Error()})
		}
	}

	if len(tasks) == 0 && len(discoverFails) == 0 {
		result.Warnings = append(result.Warnings, "未发现可转换的 Markdown 文件")
	}

	result.FailureCount = len(result.Failures)
	result.WarningCount = len(result.Warnings)
	return result, nil
}
