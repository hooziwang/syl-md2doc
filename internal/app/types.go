package app

import (
	"syl-md2doc/internal/convert"
)

type Options struct {
	Inputs         []string
	OutputArg      string
	Jobs           int
	ReferenceDocx  string
	PandocPath     string
	HighlightWords []string
	CWD            string
	Verbose        bool
	Converter      convert.Converter
}

type Failure struct {
	Source string
	Reason string
}

type Result struct {
	SuccessCount int
	FailureCount int
	WarningCount int
	Warnings     []string
	Failures     []Failure
	OutputPaths  []string
	PandocPath   string
	PandocVer    string
}
