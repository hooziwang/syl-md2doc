APP := syl-md2doc
GO ?= go
BIN_DIR ?= bin
BIN := $(BIN_DIR)/$(APP)
DESTDIR ?=
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -X 'syl-md2doc/cmd.Version=$(VERSION)' -X 'syl-md2doc/cmd.Commit=$(COMMIT)' -X 'syl-md2doc/cmd.BuildTime=$(BUILD_TIME)'
GO_BIN_DIR ?= $(shell sh -c 'gobin="$$( $(GO) env GOBIN )"; if [ -n "$$gobin" ]; then printf "%s" "$$gobin"; else gopath="$$( $(GO) env GOPATH )"; printf "%s/bin" "$${gopath%%:*}"; fi')
INSTALL_BIN_DIR := $(DESTDIR)$(GO_BIN_DIR)
INSTALL_BIN := $(INSTALL_BIN_DIR)/$(APP)
DEFAULT_GOAL := default

INPUTS ?=
OUTPUT ?=
REFERENCE_DOCX ?=
PANDOC_PATH ?=
JOBS ?=
VERBOSE ?=

.DEFAULT_GOAL := $(DEFAULT_GOAL)

.PHONY: default help build test fmt tidy clean run install uninstall

default:
	@$(MAKE) fmt
	@$(MAKE) test
	@$(MAKE) install

help:
	@echo "Targets:"
	@echo "  make              - 默认流程：fmt -> test -> install"
	@echo "  make build        - 编译二进制到 $(BIN)"
	@echo "  make test         - 运行全部测试"
	@echo "  make fmt          - gofmt 全部 Go 文件"
	@echo "  make tidy         - 整理 go.mod/go.sum"
	@echo "  make run          - 直跑入口（需要 INPUTS）"
	@echo "  make install      - 安装到 Go bin 目录（GOBIN 或 GOPATH/bin）"
	@echo "  make uninstall    - 卸载已安装二进制"
	@echo "  make clean        - 删除构建产物"
	@echo ""
	@echo "Variables:"
	@echo "  INPUTS='a.md dir1'        必填（run），支持多个输入"
	@echo "  OUTPUT=/path/out          可选（目录或 docx 文件）"
	@echo "  REFERENCE_DOCX=/path/ref.docx  可选"
	@echo "  PANDOC_PATH=/path/pandoc  可选"
	@echo "  JOBS=8                    可选"
	@echo "  VERBOSE=1                可选（1 表示 --verbose）"
	@echo "  GO_BIN_DIR=...            覆盖安装目录（默认 GOBIN 或 GOPATH/bin）"
	@echo "  DESTDIR=                  打包场景根目录"
	@echo "  VERSION=v0.1.0            可选，覆盖版本号"
	@echo "  COMMIT=abc1234            可选，覆盖提交哈希"
	@echo "  BUILD_TIME=...            可选，覆盖构建时间（UTC）"

build:
	@mkdir -p $(BIN_DIR)
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BIN) .

test:
	$(GO) test ./...

fmt:
	@gofmt -w $$(find . -name '*.go' -type f)

tidy:
	$(GO) mod tidy

run:
	@if [ -z "$(INPUTS)" ]; then echo "还没传 INPUTS（至少一个文件或目录）"; exit 1; fi
	$(GO) run . $(INPUTS) \
		$(if $(OUTPUT),--output "$(OUTPUT)",) \
		$(if $(REFERENCE_DOCX),--reference-docx "$(REFERENCE_DOCX)",) \
		$(if $(PANDOC_PATH),--pandoc-path "$(PANDOC_PATH)",) \
		$(if $(JOBS),--jobs $(JOBS),) \
		$(if $(filter 1 true TRUE yes YES,$(VERBOSE)),--verbose,)

clean:
	rm -rf $(BIN_DIR)

install: build
	@mkdir -p "$(INSTALL_BIN_DIR)"
	install -m 0755 "$(BIN)" "$(INSTALL_BIN)"

uninstall:
	rm -f "$(INSTALL_BIN)"
