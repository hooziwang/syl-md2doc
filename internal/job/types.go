package job

type Task struct {
	SourcePath string
	TargetPath string
}

type Result struct {
	Task     Task
	Warnings []string
	Error    error
}
