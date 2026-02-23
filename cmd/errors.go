package cmd

import "errors"

var errBuildFailed = errors.New("存在转换失败项")

func IsReportedError(err error) bool {
	return errors.Is(err, errBuildFailed)
}
