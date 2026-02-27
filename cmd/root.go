package cmd

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"syl-md2doc/internal/app"
)

type buildFlags struct {
	outputArg      string
	jobs           int
	referenceDocx  string
	pandocPath     string
	highlightWords string
	verbose        bool
}

const rootLongHelp = `将一个或多个 Markdown 文件批量转换为 Word(.docx)。

输入规则：
1. 支持多个文件、多个目录、文件与目录混合输入。
2. 目录会递归扫描；仅处理 .md 文件，其他文件自动忽略。
3. 一个 .md 文件对应一个 .docx 文件。

输出规则：
1. 默认输出到当前目录。
2. 目录输入会保留相对路径结构。
3. 生成文件名会追加 6 位字母数字识别码（如 a_Xy12Z9.docx）；冲突时自动重生识别码。
4. 成功时默认输出精简 summary；失败时输出详细诊断与修复建议。

依赖规则：
1. 依赖 pandoc 完成转换。
2. 可用 --pandoc-path 指定 pandoc 绝对路径。
3. 建议使用较新版本 pandoc（如 >= 2.19.0）。
4. 可用 --highlight-words/-w 指定需要高亮的词；支持英文/中文逗号、英文/中文分号、空格或换行分隔（需模板中存在 KeywordHighlight 字符样式）。`

const rootExamples = `  # 单文件转换（输出到当前目录）
  syl-md2doc /abs/docs/a.md

  # 多输入（文件 + 目录）
  syl-md2doc /abs/docs/a.md /abs/docs/chapter

  # 指定输出目录
  syl-md2doc /abs/docs/chapter --output /abs/out

  # 单输入时指定输出文件
  syl-md2doc /abs/docs/a.md --output /abs/out/final.docx

  # 指定模板与 pandoc 路径（建议使用绝对路径）
  syl-md2doc /abs/docs/chapter --reference-docx /abs/template/ref.docx --pandoc-path /abs/bin/pandoc

  # 指定需要高亮的词（支持逗号/分号/空格/换行）
  syl-md2doc /abs/docs/a.md -w "paper,lanterns,classroom"

  # 查看版本（兼容两种写法）
  syl-md2doc --version
  syl-md2doc version`

func Execute() error {
	root := NewRootCmd(os.Stdout, os.Stderr)
	root.SetArgs(normalizeArgs(os.Args[1:]))
	return root.Execute()
}

func NewRootCmd(stdout io.Writer, stderr io.Writer) *cobra.Command {
	flags := &buildFlags{}
	showVersion := false

	root := &cobra.Command{
		Use:           "syl-md2doc [inputs...]",
		Short:         "批量将 Markdown 转为 docx",
		Long:          rootLongHelp,
		Example:       rootExamples,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          runBuild(stdout, stderr, flags, &showVersion),
	}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.CompletionOptions.HiddenDefaultCmd = true
	bindBuildFlags(root, flags)
	root.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "显示版本信息")
	return root
}

func bindBuildFlags(cmd *cobra.Command, flags *buildFlags) {
	cmd.PersistentFlags().StringVarP(&flags.outputArg, "output", "o", "", "输出目录或输出文件")
	cmd.PersistentFlags().IntVarP(&flags.jobs, "jobs", "j", runtime.NumCPU(), "并发任务数")
	cmd.PersistentFlags().StringVar(&flags.referenceDocx, "reference-docx", "", "pandoc 参考 docx 模板")
	cmd.PersistentFlags().StringVar(&flags.pandocPath, "pandoc-path", "", "pandoc 可执行文件路径")
	cmd.PersistentFlags().StringVarP(&flags.highlightWords, "highlight-words", "w", "", "需要高亮的词，支持逗号/分号/空格/换行分隔")
	cmd.PersistentFlags().BoolVar(&flags.verbose, "verbose", false, "输出详细日志")
}

func runBuild(stdout io.Writer, stderr io.Writer, flags *buildFlags, showVersion *bool) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if showVersion != nil && *showVersion {
			printVersion(stdout)
			return nil
		}
		if len(args) == 0 {
			emitNDJSON(stderr, "error", "invalid_input", "缺少输入参数", map[string]any{
				"required": "至少一个 .md 文件或目录",
				"args":     args,
			}, suggestionForTopError("至少提供一个输入"))
			return errBuildFailed
		}

		cwd, err := os.Getwd()
		if err != nil {
			emitNDJSON(stderr, "error", "cwd_read_failed", "读取当前目录失败", map[string]any{
				"error": err.Error(),
			}, "检查运行目录是否可访问，或在可访问目录中重试")
			return errBuildFailed
		}
		start := time.Now()
		highlightWords := parseHighlightWords(flags.highlightWords)
		if flags.verbose {
			emitNDJSON(stdout, "info", "build_start", "开始执行 Markdown 转 docx", map[string]any{
				"cwd":             cwd,
				"inputs":          absPaths(cwd, args),
				"output_arg":      absPath(cwd, flags.outputArg),
				"jobs":            flags.jobs,
				"reference_docx":  absPath(cwd, flags.referenceDocx),
				"pandoc_path":     absPath(cwd, flags.pandocPath),
				"highlight_words": highlightWords,
				"verbose":         flags.verbose,
			}, "")
		}

		if flags.verbose && flags.referenceDocx != "" && !filepath.IsAbs(flags.referenceDocx) {
			emitNDJSON(stdout, "info", "reference_docx_resolved", "已解析 reference-docx 绝对路径", map[string]any{
				"raw":      flags.referenceDocx,
				"resolved": absPath(cwd, flags.referenceDocx),
			}, "")
		}

		res, err := app.Run(app.Options{
			Inputs:         args,
			OutputArg:      flags.outputArg,
			Jobs:           flags.jobs,
			ReferenceDocx:  flags.referenceDocx,
			PandocPath:     flags.pandocPath,
			HighlightWords: highlightWords,
			CWD:            cwd,
			Verbose:        flags.verbose,
		})
		if err != nil {
			emitNDJSON(stderr, "error", "build_aborted", "转换任务启动失败", map[string]any{
				"error":  err.Error(),
				"inputs": absPaths(cwd, args),
			}, suggestionForTopError(err.Error()))
			return errBuildFailed
		}
		if flags.verbose {
			emitNDJSON(stdout, "info", "pandoc_environment", "pandoc 环境检测结果", map[string]any{
				"pandoc_path":    absPath(cwd, res.PandocPath),
				"pandoc_version": res.PandocVer,
			}, "")
		}

		// 成功场景默认精简输出；失败或 --verbose 时输出逐条告警。
		if flags.verbose || res.FailureCount > 0 {
			for idx, w := range res.Warnings {
				emitNDJSON(stderr, "warn", "warning", "处理过程中产生告警", map[string]any{
					"index":   idx + 1,
					"warning": w,
				}, "根据 warning 内容检查资源路径、文件格式或输入范围")
			}
		}
		for idx, f := range res.Failures {
			emitNDJSON(stderr, "error", "file_failed", "文件转换失败", map[string]any{
				"index":       idx + 1,
				"source_path": absPath(cwd, f.Source),
				"reason":      f.Reason,
			}, suggestionForFailure(f.Reason))
		}

		level := "info"
		status := "success"
		if res.FailureCount > 0 {
			level = "error"
			status = "partial_failed"
		}
		summaryDetails := map[string]any{
			"status":        status,
			"success_count": res.SuccessCount,
			"failure_count": res.FailureCount,
			"warning_count": res.WarningCount,
			"duration_ms":   time.Since(start).Milliseconds(),
			"output_paths":  res.OutputPaths,
		}
		if len(res.OutputPaths) == 1 {
			summaryDetails["output_path"] = res.OutputPaths[0]
		}
		// 失败时给完整诊断上下文；成功默认只保留结果导向字段。
		if res.FailureCount > 0 || flags.verbose {
			summaryDetails["pandoc_path"] = absPath(cwd, res.PandocPath)
			summaryDetails["pandoc_version"] = res.PandocVer
			summaryDetails["inputs"] = absPaths(cwd, args)
			summaryDetails["output_arg"] = absPath(cwd, flags.outputArg)
			summaryDetails["jobs"] = flags.jobs
		}
		suggestion := ""
		if res.FailureCount > 0 {
			suggestion = "修复失败项后重试；建议先按 file_failed 事件逐项处理"
		}
		emitNDJSON(stdout, level, "summary", "批量转换完成", summaryDetails, suggestion)
		if res.FailureCount > 0 {
			return errBuildFailed
		}
		_ = cmd
		return nil
	}
}

func normalizeArgs(args []string) []string {
	if len(args) == 1 && args[0] == "version" {
		return []string{"--version"}
	}
	return args
}

func parseHighlightWords(raw string) []string {
	normalized := strings.NewReplacer("，", ",", ";", ",", "；", ",", "\n", ",", "\r", ",").Replace(raw)
	parts := strings.Split(normalized, ",")
	seen := make(map[string]struct{})
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		for _, token := range strings.Fields(strings.TrimSpace(part)) {
			word := strings.ToLower(strings.TrimSpace(token))
			if word == "" {
				continue
			}
			if _, ok := seen[word]; ok {
				continue
			}
			seen[word] = struct{}{}
			out = append(out, word)
		}
	}
	return out
}
