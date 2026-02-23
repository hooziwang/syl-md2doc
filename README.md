# syl-md2doc

将一个或多个 Markdown 文件批量转换为 Word (`.docx`)。

## 特性

- 支持输入多个文件、多个目录、文件与目录混合。
- 目录输入递归扫描；仅处理 `.md` 文件。
- 一个 `.md` 对应一个 `.docx`。
- Markdown 解析使用 `CommonMark + GFM`（通过 `pandoc`）。
- 支持 `--reference-docx` 控制最终 Word 样式。
- 批量执行时单文件失败不中断，最终汇总失败并返回非 0。
- 同名输出自动追加 `_1/_2/...`，不会覆盖已有文件。

## 前置依赖

需要预先安装 `pandoc`。

- macOS: `brew install pandoc`
- Ubuntu/Debian: `sudo apt-get install pandoc`
- Windows (Scoop): `scoop install pandoc`

如果 `pandoc` 不在 PATH，可使用 `--pandoc-path` 指定。

## 用法

### 入口 1（直跑）

```bash
syl-md2doc <inputs...> [--output ...] [--jobs ...] [--reference-docx ...]
```

### 入口 2（子命令）

```bash
syl-md2doc build <inputs...> [--output ...] [--jobs ...] [--reference-docx ...]
```

### 版本

```bash
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
- `--pandoc-path`: pandoc 可执行文件路径。
- `--verbose`: 打印更详细执行信息。

## 输出规则

- 目录输入：在输出目录下保留相对路径结构。
- 单独文件输入：输出到输出根目录。
- 非 `.md` 输入：忽略并输出 `warn`。
- 本地图片缺失：记录告警并继续（若 pandoc 仍产出 docx）。

## 退出码

- 全部成功：`0`
- 存在失败项：`1`

