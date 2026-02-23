package input

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func Discover(inputs []string, cwd string) ([]SourceItem, []string, []Failure, error) {
	if strings.TrimSpace(cwd) == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("读取当前目录失败：%w", err)
		}
		cwd = wd
	}

	items := make([]SourceItem, 0)
	warns := make([]string, 0)
	fails := make([]Failure, 0)

	for _, raw := range inputs {
		in := strings.TrimSpace(raw)
		if in == "" {
			continue
		}
		abs := in
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(cwd, abs)
		}
		abs = filepath.Clean(abs)

		st, err := os.Stat(abs)
		if err != nil {
			fails = append(fails, Failure{Input: abs, Reason: "输入不存在或不可访问"})
			continue
		}

		if st.IsDir() {
			walkErr := filepath.WalkDir(abs, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					warns = append(warns, fmt.Sprintf("扫描失败（已跳过）：%s", path))
					return nil
				}
				if d.IsDir() {
					return nil
				}
				if strings.EqualFold(filepath.Ext(path), ".md") {
					rel, relErr := filepath.Rel(abs, path)
					if relErr != nil {
						warns = append(warns, fmt.Sprintf("路径计算失败（已跳过）：%s", path))
						return nil
					}
					items = append(items, SourceItem{
						SourcePath: path,
						FromDir:    true,
						BaseDir:    abs,
						RelPath:    rel,
					})
					return nil
				}
				warns = append(warns, fmt.Sprintf("忽略非 Markdown 文件：%s", path))
				return nil
			})
			if walkErr != nil {
				warns = append(warns, fmt.Sprintf("目录扫描异常：%s", abs))
			}
			continue
		}

		if strings.EqualFold(filepath.Ext(abs), ".md") {
			items = append(items, SourceItem{SourcePath: abs})
			continue
		}
		warns = append(warns, fmt.Sprintf("忽略非 Markdown 文件：%s", abs))
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].SourcePath != items[j].SourcePath {
			return items[i].SourcePath < items[j].SourcePath
		}
		return items[i].RelPath < items[j].RelPath
	})
	sort.Strings(warns)
	sort.Slice(fails, func(i, j int) bool {
		return fails[i].Input < fails[j].Input
	})
	return items, warns, fails, nil
}
