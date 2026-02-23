package plan

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"syl-md2doc/internal/input"
	"syl-md2doc/internal/job"
)

type Options struct {
	OutputArg string
	CWD       string
}

func BuildTargets(sources []input.SourceItem, opts Options) ([]job.Task, []string, error) {
	if len(sources) == 0 {
		return nil, nil, nil
	}
	cwd := opts.CWD
	if strings.TrimSpace(cwd) == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, nil, fmt.Errorf("读取当前目录失败：%w", err)
		}
		cwd = wd
	}

	warns := make([]string, 0)
	outputArg := strings.TrimSpace(opts.OutputArg)
	multi := len(sources) > 1
	useFixedOutput := false
	fixedOutput := ""
	outputRoot := cwd

	if outputArg != "" {
		absOut := outputArg
		if !filepath.IsAbs(absOut) {
			absOut = filepath.Join(cwd, absOut)
		}
		absOut = filepath.Clean(absOut)

		if strings.EqualFold(filepath.Ext(absOut), ".docx") {
			if multi {
				outputRoot = filepath.Dir(absOut)
				warns = append(warns, fmt.Sprintf("多输入场景下 --output=%s 被视为目录模式（使用其父目录）", outputArg))
			} else {
				useFixedOutput = true
				fixedOutput = absOut
			}
		} else {
			outputRoot = absOut
		}
	}

	used := make(map[string]struct{}, len(sources))
	tasks := make([]job.Task, 0, len(sources))
	for i, src := range sources {
		target := ""
		if useFixedOutput && i == 0 {
			target = fixedOutput
		} else {
			if src.FromDir {
				target = filepath.Join(outputRoot, replaceExt(src.RelPath, ".docx"))
			} else {
				target = filepath.Join(outputRoot, replaceExt(filepath.Base(src.SourcePath), ".docx"))
			}
		}
		target = uniqueTarget(target, used)
		tasks = append(tasks, job.Task{SourcePath: src.SourcePath, TargetPath: target})
	}
	return tasks, warns, nil
}

func replaceExt(name, ext string) string {
	baseExt := filepath.Ext(name)
	if baseExt == "" {
		return name + ext
	}
	return strings.TrimSuffix(name, baseExt) + ext
}

func uniqueTarget(candidate string, used map[string]struct{}) string {
	candidate = filepath.Clean(candidate)
	ext := filepath.Ext(candidate)
	stem := strings.TrimSuffix(candidate, ext)
	idx := 0
	for {
		tryPath := candidate
		if idx > 0 {
			tryPath = fmt.Sprintf("%s_%d%s", stem, idx, ext)
		}
		if _, ok := used[tryPath]; ok {
			idx++
			continue
		}
		if _, err := os.Stat(tryPath); err == nil {
			idx++
			continue
		}
		used[tryPath] = struct{}{}
		return tryPath
	}
}
