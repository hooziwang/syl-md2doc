package input

type SourceItem struct {
	SourcePath string
	FromDir    bool
	BaseDir    string
	RelPath    string
}

type Failure struct {
	Input  string
	Reason string
}
