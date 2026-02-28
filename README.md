# syl-md2doc

将一个或多个 Markdown 文件批量转换为 Word (`.docx`)。

## 特性

- 支持输入多个文件、多个目录、文件与目录混合。
- 目录输入递归扫描；仅处理 `.md` 文件。
- 一个 `.md` 对应一个 `.docx`。
- Markdown 解析使用 `CommonMark + GFM`（通过 `pandoc`）。
- 支持 `--reference-docx` 控制最终 Word 样式；未指定时自动使用内置默认模板。
- 批量执行时单文件失败不中断，最终汇总失败并返回非 0。
- 生成文件名自动追加 6 位字母数字识别码（如 `listing_for_test_Xy12Z9.docx`），冲突时自动重生识别码。

## 安装

### macOS（Homebrew）

安装（首次/已 tap 过都可用）：

```bash
brew update && brew install hooziwang/tap/syl-md2doc
```

升级：

```bash
brew update && brew upgrade hooziwang/tap/syl-md2doc
```

如果提示 `No available formula`（本地 tap 索引过期）：

```bash
brew untap hooziwang/tap && brew install hooziwang/tap/syl-md2doc
```

### Windows（Scoop）

安装：

```powershell
scoop update; scoop bucket add hooziwang https://github.com/hooziwang/scoop-bucket.git; scoop install syl-md2doc
```

升级：

```powershell
scoop update; scoop update syl-md2doc
```

如果提示找不到应用（bucket 索引过期）：

```powershell
scoop bucket rm hooziwang; scoop bucket add hooziwang https://github.com/hooziwang/scoop-bucket.git; scoop update; scoop install syl-md2doc
```

## 前置依赖

程序依赖 `pandoc` 执行 Markdown 到 docx 的转换，需要预先安装：

- macOS: `brew install pandoc`
- Ubuntu/Debian: `sudo apt-get install pandoc`
- Windows (Scoop): `scoop install pandoc`

如果 `pandoc` 不在 PATH，可使用 `--pandoc-path` 指定。

运行时会做环境检查：
- 检查 `pandoc` 是否存在（PATH 或 `--pandoc-path`）。
- 读取 `pandoc` 版本信息（用于诊断，不作为阻断条件）。
- 开启 `--verbose` 时会输出检测到的 `pandoc` 路径和版本。

建议使用较新版本的 pandoc（如 `>= 2.19.0`），以获得更稳定的 Markdown 兼容性。

## 用法

### 入口（直跑）

```bash
syl-md2doc <inputs...> [--output ...] [--jobs ...] [--reference-docx ...]
```

### 版本

```bash
syl-md2doc --version
# 兼容：
syl-md2doc version
```

## 参数

- `inputs...`: 必填，文件/目录均可。
- `--output, -o`: 输出目录或输出文件。
  - 单输入时可为 `.docx` 文件路径。
  - 多输入时若是 `.docx` 文件路径，会自动按目录模式处理（使用其父目录）并告警。
  - 默认当前目录。
- `--jobs, -j`: 并发数，默认 CPU 核数。
- `--reference-docx`: Word 模板文件（传给 pandoc `--reference-doc`）。
  - 未指定时，程序会自动使用内置默认模板（已编译进二进制）。
- `--pandoc-path`: pandoc 可执行文件路径。
- 高亮约定：Markdown 中的 `**...**` 在输出 Word 时会同时应用“加粗 + `KeywordHighlight` 字符样式”。
  - 若使用自定义 `--reference-docx`，请在模板中创建 `KeywordHighlight` 字符样式并设置高亮颜色。
- `--verbose`: 打印更详细执行信息。

## 输出规则

- 目录输入：在输出目录下保留相对路径结构。
- 单独文件输入：输出到输出根目录。
- 默认生成文件名：`原文件名_6位字母数字识别码.docx`。
- 非 `.md` 输入：忽略并输出 `warn`。
- 本地图片缺失：记录告警并继续（若 pandoc 仍产出 docx）。

## 输出格式（AI 友好）

所有运行日志均为 **NDJSON**（每行一个 JSON 对象），便于 AI 与程序解析。
例外：`--version`/`-v` 输出为纯文本（版本信息 + daddylovesyl 横幅），与 `syl-md2ppt` 保持一致。

统一字段：
- `timestamp`: RFC3339Nano 时间戳（UTC）
- `level`: `info` / `warn` / `error`
- `event`: 事件名（如 `build_start`、`file_failed`、`summary`）
- `message`: 人类可读说明
- `details`: 结构化细节（尽可能使用绝对路径）
- `suggestion`: 出错或告警时的建议修复方案

输出策略：
- 成功（默认）：仅输出一条 `summary`（结果导向、简洁）。
- 成功 + `--verbose`：额外输出 `build_start`、`pandoc_environment`、逐条 `warning`。
- 失败：输出 `file_failed`（可多条）+ 一条带建议的 `summary`。

`summary.details` 关键字段：
- `status`: `success` / `partial_failed`
- `success_count`: 成功文件数
- `failure_count`: 失败文件数
- `warning_count`: 告警数
- `duration_ms`: 执行耗时（毫秒）
- `output_paths`: 成功产物绝对路径数组
- `output_path`: 当仅生成一个文件时提供（绝对路径）
- 失败或 `--verbose` 时附加：`inputs`、`output_arg`、`jobs`、`pandoc_path`、`pandoc_version`

示例：

```json
{"timestamp":"2026-02-23T10:00:01Z","level":"info","event":"summary","message":"批量转换完成","details":{"status":"success","success_count":1,"failure_count":0,"warning_count":0,"duration_ms":271,"output_path":"/abs/out/a.docx","output_paths":["/abs/out/a.docx"]}}
{"timestamp":"2026-02-23T10:00:01Z","level":"error","event":"file_failed","message":"文件转换失败","details":{"source_path":"/abs/a.md","reason":"pandoc 转换失败：..."},"suggestion":"建议先手工执行 pandoc 命令定位具体语法/资源问题，再修复 Markdown 后重试"}
{"timestamp":"2026-02-23T10:00:01Z","level":"error","event":"summary","message":"批量转换完成","details":{"status":"partial_failed","success_count":0,"failure_count":1,"warning_count":0,"duration_ms":312,"output_paths":[],"inputs":["/abs/a.md"],"pandoc_path":"/opt/homebrew/bin/pandoc","pandoc_version":"3.9.0"},"suggestion":"修复失败项后重试；建议先按 file_failed 事件逐项处理"}
```

## 常见错误与处理

- `event=invalid_input`
  - 原因：未传入任何输入文件/目录。
  - 处理：传入至少一个 `.md` 文件或目录（建议使用绝对路径）。
- `event=build_aborted` 且提示未找到 `pandoc`
  - 原因：系统未安装 pandoc 或不在 PATH。
  - 处理：安装 pandoc，或使用 `--pandoc-path /abs/path/to/pandoc`。
- `event=file_failed` 且 `reason` 包含资源缺失
  - 原因：Markdown 中引用的本地图片/资源不存在。
  - 处理：修正资源路径或补齐资源文件后重试。
- `event=file_failed` 且 `reason` 包含权限问题
  - 原因：输入不可读或输出目录不可写。
  - 处理：修正文件权限，或切换到有权限的目录。

## 常用命令示例

```bash
# 单文件转换（默认输出到当前目录）
syl-md2doc /abs/docs/a.md

# 多输入（文件 + 目录）
syl-md2doc /abs/docs/a.md /abs/docs/chapter

# 指定输出目录
syl-md2doc /abs/docs/chapter --output /abs/out

# 单输入时指定输出文件
syl-md2doc /abs/docs/a.md --output /abs/out/final.docx

# 指定 reference docx 模板
syl-md2doc /abs/docs/chapter --reference-docx /abs/template/reference.docx

# 使用 **...** 标注“加粗 + 高亮”
syl-md2doc /abs/docs/a.md

# 指定 pandoc 路径 + 详细日志
syl-md2doc /abs/docs/chapter --pandoc-path /abs/bin/pandoc --verbose
```

## 退出码

- 全部成功：`0`
- 存在失败项：`1`
