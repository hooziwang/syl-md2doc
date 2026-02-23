# syl-md2doc Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 构建一个 Go+Cobra 命令行工具，将多个 Markdown 输入批量转换为对应的 docx，并在失败可恢复的前提下输出可审计汇总。

**Architecture:** 采用分层结构：`cmd` 负责 CLI 和参数归一化，`internal/input` 负责发现与过滤，`internal/plan` 负责目标路径规划与冲突处理，`internal/convert` 封装 pandoc 转换，`internal/runner` 负责并发调度与结果汇总，`internal/app` 负责编排。通过接口隔离转换后端，首版仅落地 pandoc 适配器。

**Tech Stack:** Go 1.22+, Cobra, stretchr/testify, os/exec（调用 pandoc）

---

### Task 1: 初始化项目骨架与 CLI 最小入口

**Files:**
- Create: `go.mod`
- Create: `main.go`
- Create: `cmd/root.go`
- Create: `cmd/version.go`
- Create: `cmd/errors.go`
- Test: `cmd/root_test.go`

**Step 1: Write the failing test**

```go
func TestBuildRequiresAtLeastOneInput(t *testing.T) {
    cmd := NewRootCmd(...)
    cmd.SetArgs([]string{"build"})
    err := cmd.Execute()
    require.Error(t, err)
    require.Contains(t, err.Error(), "至少提供一个输入")
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./cmd -run TestBuildRequiresAtLeastOneInput -v`  
Expected: FAIL，提示 `NewRootCmd` 或校验逻辑未实现。

**Step 3: Write minimal implementation**

```go
root := &cobra.Command{Use: "syl-md2doc [inputs...]"}
build := &cobra.Command{Use: "build [inputs...]", Args: cobra.MinimumNArgs(1)}
```

**Step 4: Run test to verify it passes**

Run: `go test ./cmd -run TestBuildRequiresAtLeastOneInput -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add go.mod main.go cmd/root.go cmd/version.go cmd/errors.go cmd/root_test.go
git commit -m "feat(cli): 初始化 md2doc 命令入口与参数骨架"
```

### Task 2: CLI 参数与直跑归一化

**Files:**
- Modify: `cmd/root.go`
- Test: `cmd/root_test.go`

**Step 1: Write the failing tests**

```go
func TestNormalizeArgs_InsertBuildForDirectRun(t *testing.T) {}
func TestParseFlags_JobsAndOutputAndReferenceDocx(t *testing.T) {}
```

**Step 2: Run tests to verify failure**

Run: `go test ./cmd -run 'TestNormalizeArgs|TestParseFlags' -v`  
Expected: FAIL

**Step 3: Implement minimal parsing/normalization**

```go
// 非子命令输入自动补 build
if firstArgIsPositional(args) { args = append([]string{"build"}, args...) }
```

**Step 4: Run tests to verify pass**

Run: `go test ./cmd -run 'TestNormalizeArgs|TestParseFlags' -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/root.go cmd/root_test.go
git commit -m "feat(cli): 支持直跑 build 归一化及核心参数解析"
```

### Task 3: 输入发现与过滤（混合输入 + 递归）

**Files:**
- Create: `internal/input/discover.go`
- Create: `internal/input/types.go`
- Test: `internal/input/discover_test.go`

**Step 1: Write failing tests**

```go
func TestDiscover_MixedFileAndDir_RecursiveOnlyMD(t *testing.T) {}
func TestDiscover_NonMDWarnAndIgnore(t *testing.T) {}
func TestDiscover_MissingInputAsFailureItem(t *testing.T) {}
```

**Step 2: Run tests to verify failure**

Run: `go test ./internal/input -v`  
Expected: FAIL

**Step 3: Implement minimal discover logic**

```go
func Discover(inputs []string) (items []SourceItem, warns []string, fails []DiscoverFailure)
```

**Step 4: Run tests to verify pass**

Run: `go test ./internal/input -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/input/discover.go internal/input/types.go internal/input/discover_test.go
git commit -m "feat(input): 实现混合输入发现、递归扫描与 md 过滤"
```

### Task 4: 输出路径规划与冲突去重

**Files:**
- Create: `internal/plan/targets.go`
- Test: `internal/plan/targets_test.go`

**Step 1: Write failing tests**

```go
func TestPlanTargets_DirInputPreserveRelativePath(t *testing.T) {}
func TestPlanTargets_SingleFilesToOutputRoot(t *testing.T) {}
func TestPlanTargets_NameConflictAppendSuffix(t *testing.T) {}
func TestPlanTargets_MultiInputOutputFileFallbackToDirWarn(t *testing.T) {}
```

**Step 2: Run tests to verify failure**

Run: `go test ./internal/plan -v`  
Expected: FAIL

**Step 3: Implement minimal planning logic**

```go
func BuildTargets(sources []input.SourceItem, opts Options) (tasks []Task, warns []string, err error)
```

**Step 4: Run tests to verify pass**

Run: `go test ./internal/plan -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/plan/targets.go internal/plan/targets_test.go
git commit -m "feat(plan): 实现输出目标规划与同名冲突后缀策略"
```

### Task 5: Pandoc 转换器接口与实现

**Files:**
- Create: `internal/convert/converter.go`
- Create: `internal/convert/pandoc.go`
- Test: `internal/convert/pandoc_test.go`

**Step 1: Write failing tests**

```go
func TestPandocConverter_CommandBuild(t *testing.T) {}
func TestPandocConverter_MissingBinary(t *testing.T) {}
func TestPandocConverter_MissingImageClassifiedAsWarning(t *testing.T) {}
```

**Step 2: Run tests to verify failure**

Run: `go test ./internal/convert -v`  
Expected: FAIL

**Step 3: Implement converter and stderr classification**

```go
type Converter interface { Convert(ctx context.Context, task Task) Result }
```

**Step 4: Run tests to verify pass**

Run: `go test ./internal/convert -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/convert/converter.go internal/convert/pandoc.go internal/convert/pandoc_test.go
git commit -m "feat(convert): 实现 pandoc 转换器与缺图告警分类"
```

### Task 6: 并发执行器与结果汇总

**Files:**
- Create: `internal/runner/runner.go`
- Create: `internal/runner/types.go`
- Test: `internal/runner/runner_test.go`

**Step 1: Write failing tests**

```go
func TestRunner_ContinueOnTaskFailure(t *testing.T) {}
func TestRunner_RespectJobsLimit(t *testing.T) {}
func TestRunner_CollectSummaryCounts(t *testing.T) {}
```

**Step 2: Run tests to verify failure**

Run: `go test ./internal/runner -v`  
Expected: FAIL

**Step 3: Implement worker pool**

```go
func Run(ctx context.Context, jobs int, tasks []Task, c convert.Converter) Summary
```

**Step 4: Run tests to verify pass**

Run: `go test ./internal/runner -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/runner/runner.go internal/runner/types.go internal/runner/runner_test.go
git commit -m "feat(runner): 实现并发执行与失败不中断汇总"
```

### Task 7: 应用编排层（app）

**Files:**
- Create: `internal/app/run.go`
- Create: `internal/app/types.go`
- Test: `internal/app/run_test.go`

**Step 1: Write failing tests**

```go
func TestAppRun_EndToEndSummaryAndExitBehavior(t *testing.T) {}
func TestAppRun_PandocUnavailableReturnActionableError(t *testing.T) {}
```

**Step 2: Run tests to verify failure**

Run: `go test ./internal/app -v`  
Expected: FAIL

**Step 3: Implement orchestration**

```go
// Discover -> Plan -> Runner -> Summary
func Run(opts Options) (Result, error)
```

**Step 4: Run tests to verify pass**

Run: `go test ./internal/app -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add internal/app/run.go internal/app/types.go internal/app/run_test.go
git commit -m "feat(app): 串联发现、规划、转换与结果汇总流程"
```

### Task 8: CLI 集成 app 并输出用户可读汇总

**Files:**
- Modify: `cmd/root.go`
- Test: `cmd/root_test.go`

**Step 1: Write failing tests**

```go
func TestCLI_BuildPrintsSummaryAndReturnsNonZeroOnFailures(t *testing.T) {}
func TestCLI_VersionCommand(t *testing.T) {}
```

**Step 2: Run tests to verify failure**

Run: `go test ./cmd -v`  
Expected: FAIL

**Step 3: Implement summary output and exit policy**

```go
if result.FailureCount > 0 { return errBuildFailed }
```

**Step 4: Run tests to verify pass**

Run: `go test ./cmd -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add cmd/root.go cmd/root_test.go
git commit -m "feat(cli): 输出转换汇总并按失败数控制退出码"
```

### Task 9: 端到端样例与回归测试

**Files:**
- Create: `testdata/cases/basic/input/*.md`
- Create: `testdata/cases/basic/reference.docx` (optional fixture)
- Create: `internal/app/e2e_test.go`

**Step 1: Write failing e2e test**

```go
func TestE2E_BatchConvertWithMixedInputs(t *testing.T) {}
```

**Step 2: Run test to verify failure**

Run: `go test ./internal/app -run TestE2E_BatchConvertWithMixedInputs -v`  
Expected: FAIL

**Step 3: Implement missing glue code**

```go
// 补全路径清洗、默认输出目录、后缀冲突处理遗漏分支
```

**Step 4: Run test to verify pass**

Run: `go test ./internal/app -run TestE2E_BatchConvertWithMixedInputs -v`  
Expected: PASS

**Step 5: Commit**

```bash
git add testdata internal/app/e2e_test.go
git commit -m "test(e2e): 覆盖混合输入批量转换回归场景"
```

### Task 10: 文档与发布入口完善

**Files:**
- Create: `README.md`
- Create: `Makefile`
- Modify: `main.go` (if needed)

**Step 1: Write doc checks (manual checklist in plan)**

```text
- README 包含安装 pandoc 的三平台说明
- README 包含常见错误排查
- README 包含 5 个核心用例命令
```

**Step 2: Run full verification**

Run: `go test ./...`  
Expected: PASS

**Step 3: Verify build**

Run: `go build ./...`  
Expected: PASS

**Step 4: Final polish**

```text
- 精简输出文案
- 保证错误信息可操作
```

**Step 5: Commit**

```bash
git add README.md Makefile main.go
git commit -m "docs: 补充使用说明与环境依赖排查指南"
```

