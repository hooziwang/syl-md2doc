package cmd

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"syl-md2doc/internal/app"
)

type buildFlags struct {
	outputArg     string
	jobs          int
	referenceDocx string
	pandocPath    string
	verbose       bool
}

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
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          runBuild(stdout, stderr, flags, false, &showVersion),
	}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.CompletionOptions.HiddenDefaultCmd = true
	bindBuildFlags(root, flags)
	root.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "显示版本信息")

	buildCmd := &cobra.Command{
		Use:           "build [inputs...]",
		Short:         "执行 Markdown 到 docx 的批量转换",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          runBuild(stdout, stderr, flags, true, &showVersion),
	}
	root.AddCommand(buildCmd)

	versionCmd := &cobra.Command{
		Use:           "version",
		Short:         "显示版本信息",
		SilenceUsage:  true,
		SilenceErrors: true,
		Run: func(cmd *cobra.Command, args []string) {
			printVersion(stdout)
		},
	}
	root.AddCommand(versionCmd)
	return root
}

func bindBuildFlags(cmd *cobra.Command, flags *buildFlags) {
	cmd.PersistentFlags().StringVarP(&flags.outputArg, "output", "o", "", "输出目录或输出文件")
	cmd.PersistentFlags().IntVarP(&flags.jobs, "jobs", "j", runtime.NumCPU(), "并发任务数")
	cmd.PersistentFlags().StringVar(&flags.referenceDocx, "reference-docx", "", "pandoc 参考 docx 模板")
	cmd.PersistentFlags().StringVar(&flags.pandocPath, "pandoc-path", "", "pandoc 可执行文件路径")
	cmd.PersistentFlags().BoolVar(&flags.verbose, "verbose", false, "输出详细日志")
}

func runBuild(stdout io.Writer, stderr io.Writer, flags *buildFlags, subcommand bool, showVersion *bool) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if showVersion != nil && *showVersion {
			printVersion(stdout)
			return nil
		}
		if len(args) == 0 {
			if !subcommand {
				fmt.Fprintln(stderr, "至少提供一个输入（文件或目录）")
				return fmt.Errorf("至少提供一个输入")
			}
			_ = cmd.Help()
			return fmt.Errorf("至少提供一个输入")
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("读取当前目录失败：%w", err)
		}

		res, err := app.Run(app.Options{
			Inputs:        args,
			OutputArg:     flags.outputArg,
			Jobs:          flags.jobs,
			ReferenceDocx: flags.referenceDocx,
			PandocPath:    flags.pandocPath,
			CWD:           cwd,
			Verbose:       flags.verbose,
		})
		if err != nil {
			return err
		}

		for _, w := range res.Warnings {
			fmt.Fprintf(stderr, "warn: %s\n", w)
		}
		for _, f := range res.Failures {
			fmt.Fprintf(stderr, "fail: %s -> %s\n", f.Source, f.Reason)
		}

		fmt.Fprintf(stdout, "完成：成功 %d，失败 %d，告警 %d\n", res.SuccessCount, res.FailureCount, res.WarningCount)
		if res.FailureCount > 0 {
			return errBuildFailed
		}
		return nil
	}
}

func normalizeArgs(args []string) []string {
	if len(args) == 0 {
		return args
	}
	first := args[0]
	switch first {
	case "build", "help", "completion", "version":
		return args
	}
	if first == "-h" || first == "--help" || first == "-v" || first == "--version" {
		return args
	}
	if !containsPositionalInput(args) {
		return args
	}
	return append([]string{"build"}, args...)
}

func containsPositionalInput(args []string) bool {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			return i+1 < len(args)
		}
		if arg == "--output" || arg == "--jobs" || arg == "--reference-docx" || arg == "--pandoc-path" {
			i++
			continue
		}
		if strings.HasPrefix(arg, "--output=") || strings.HasPrefix(arg, "--jobs=") || strings.HasPrefix(arg, "--reference-docx=") || strings.HasPrefix(arg, "--pandoc-path=") {
			continue
		}
		if strings.HasPrefix(arg, "-") {
			continue
		}
		return true
	}
	return false
}
