package plan

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"syl-md2doc/internal/input"
	"syl-md2doc/internal/job"
)

type Options struct {
	OutputArg string
	CWD       string
}

const codeAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var codeGenerator = randomCode

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
		generatedName := false
		if useFixedOutput && i == 0 {
			target = fixedOutput
		} else {
			if src.FromDir {
				target = filepath.Join(outputRoot, replaceExt(src.RelPath, ".docx"))
			} else {
				target = filepath.Join(outputRoot, replaceExt(filepath.Base(src.SourcePath), ".docx"))
			}
			generatedName = true
		}
		target = uniqueTarget(target, used, generatedName)
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

func uniqueTarget(candidate string, used map[string]struct{}, addCode bool) string {
	candidate = filepath.Clean(candidate)
	if addCode {
		for {
			tryPath := withCode(candidate, codeGenerator(6))
			if _, ok := used[tryPath]; ok {
				continue
			}
			if _, err := os.Stat(tryPath); err == nil {
				continue
			}
			used[tryPath] = struct{}{}
			return tryPath
		}
	}

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

func withCode(path, code string) string {
	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(filepath.Base(path), ext)
	name := base + "_" + code + ext
	return filepath.Join(dir, name)
}

func randomCode(n int) string {
	if n <= 0 {
		return ""
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		// 极少发生时退化为时间字节，保证流程不中断。
		ns := time.Now().UnixNano()
		for i := range buf {
			buf[i] = byte(ns >> uint((i%8)*8))
		}
	}
	out := make([]byte, n)
	for i := range buf {
		out[i] = codeAlphabet[int(buf[i])%len(codeAlphabet)]
	}
	return string(out)
}
