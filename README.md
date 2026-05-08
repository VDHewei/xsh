# xsh

<p align="center">
  <strong>xsh</strong> - eXecute SHell tasks | 任务执行工具
</p>

<p align="center">
  English | <a href="README_CN.md">中文文档</a>
</p>

---

xsh is a Go-based task execution tool that reads task files (txt/md) and automates HTTP, SSH, gRPC task generation and execution. It features a config-driven TUI with i18n support, LLM-powered task analysis, and a structured command system.

## Features

- **Task File Parsing** - Parse `.txt`/`.md` structured task commands
- **Command Syntax** - `@{command}:{description}` style command parsing
  - `@ask:` - Interactive LLM-powered user confirmation
  - `@wait:` - Timed wait with duration
  - `@check:` - LLM-powered task result verification
- **HTTP Task Execution** - Full method coverage: `[GET]`/`[POST]`/`[PUT]`/`[PATCH]`/`[DELETE]`/`[HEAD]`/`[OPTIONS]` with headers and body
- **TUI Interactive Mode** - Terminal UI with command list, task display, progress tracking, and confirmation dialogs
  - Config-driven layout and style via `.xsh.toml` (cwd > `~/.xsh.toml` > install dir)
  - i18n support: zh-CN (default), zh-TW, en (auto-detected from system locale)
  - Custom commands from `commands/` directory (.md files with LLM parsing)
- **LLM Integration** - Local ONNX GenAI model inference for task analysis and planning
  - Real inference via `onnxruntime-genai` (tokenization, autoregressive generation, KV cache, sampling)
  - Streaming output (`-stream`)
  - Automatic fallback to regex/mock when no model is available
- **Model Management** - `model search`, `model list`, `model select` CLI commands
- **CLI Mode** - `-i` input file, `-o` output results file
- **Configurable** - `.xsh.toml` for layout, colors, UI strings, language

## Installation

```bash
go install github.com/VDHewei/xsh/cmd/xsh@latest
```

Or build from source:

```bash
git clone https://github.com/VDHewei/xsh.git
cd xsh
go build -o xsh ./cmd/xsh
```

## Usage

### Interactive TUI Mode

```bash
./xsh
```

- **Tab**: Switch focus between command list and input field
- **Enter on command**: Load command from `commands/` directory
- **Enter when tasks loaded**: Execute next task
- **Custom commands**: Place `.md` files in `commands/` directory (see [Command Format](#custom-command-format))

### CLI Mode

```bash
# Execute tasks from file
./xsh -i tasks.txt

# Execute and save results
./xsh -i tasks.txt -o results.txt

# Use LLM model for task analysis
./xsh -m models/deepseek-r1-distill -i migration.txt

# LLM inference with streaming output
./xsh -m models/deepseek-r1-distill -p "Explain this migration plan" -stream
```

### Configuration

Create `.xsh.toml` in any of these locations (first found wins):

1. Current directory: `./xsh.toml`
2. Home directory: `~/.xsh.toml`
3. Install directory: `<xsh_binary_dir>/.xsh.toml`

```toml
language = ""  # "" = auto-detect, "zh-CN", "zh-TW", "en"

[style]
header = "green"
success = "green"
error = "red"
progress = "yellow"

[layout]
command_list_width = 25

[i18n]
header = "xsh - Task Execution Tool"
task_list_title = "Task List"
progress_title = "Progress"
# ... (see .xsh.toml for all keys)
```

### Model Management

xsh supports candidate models that can be downloaded on demand:

| Model | Short Name | Default | Status |
|-------|-----------|---------|--------|
| DeepSeek R1 Distill Qwen | `deepseek` | Yes | Installed |
| GLM 5.1 Distill | `glm5.1` | No | Candidate |

```bash
# Search models on HuggingFace
./xsh model search onnx

# List local models (shows [candidate], [not installed] markers)
./xsh model list

# Select a model (use short name for candidates)
./xsh model select deepseek
./xsh model select glm5.1

# Download a candidate model
./xsh model download glm5.1
```

### Prerequisites for LLM Inference

1. **GenAI shared libraries** - Auto-downloaded to `lib/` on first use and cached for subsequent runs:
   - `onnxruntime-genai.dll` (Windows) / `.so` (Linux) / `.dylib` (macOS)
   - `onnxruntime.dll` (Windows) / `.so` (Linux) / `.dylib` (macOS)

2. **ONNX model directory** - Download from HuggingFace, e.g.:
   ```
   models/yasserrmd_deepseek-r1-distill-qwen-onnx/
   ├── genai_config.json
   ├── model.onnx
   ├── model.onnx.data
   ├── tokenizer.json
   └── tokenizer_config.json
   ```

   Use short names with `model download` to auto-resolve to the correct repo:
   ```bash
   ./xsh model download deepseek    # → yasserrmd/deepseek-r1-distill-qwen-onnx
   ./xsh model download glm5.1      # → yasserrmd/glm5.1-distill-onnx
   ```

## Task File Format

```text
# Comments start with # or >
# This is a comment
> This is also a comment

# HTTP requests (all methods supported)
[GET] http://example.com/api/health
[POST] http://example.com/api/deploy
header: Content-Type=application/json
body: {"action":"deploy"}
[PUT] http://example.com/api/update
[PATCH] http://example.com/api/patch
[DELETE] http://example.com/api/remove
[HEAD] http://example.com/api/head
[OPTIONS] http://example.com/api/options

# Interactive commands
@ask: Confirm to proceed with deployment?
@wait: 10min
@check: Verify deployment status
```

## Custom Command Format

Place `.md` files in the `commands/` directory:

```markdown
# Command Name

## 描述
Brief description of what this command does.

## 任务
[GET] http://localhost:18080/health
@ask: Confirm to proceed?
[POST] http://localhost:18080/deploy
@wait: 2min
@check: Verify deployment result
```

The `## 任务` section is parsed by LLM (with regex fallback) to extract task lines.

## Project Structure

```
xsh/
├── cmd/xsh/           # Entry point
├── commands/           # Custom command .md files
├── internal/
│   ├── config/        # Config loading, locale detection
│   ├── executor/      # Task execution engine (HTTP/SSH/gRPC/Ask/Check/Wait)
│   ├── i18n/          # i18n string manager
│   ├── parser/        # Task file parser
│   ├── tui/           # Terminal UI (tview-based)
│   └── types/         # Type definitions
├── pkg/llm/           # LLM integration
│   ├── analyzer.go    # Task analyzer
│   ├── ask_executor.go    # Ask command execution
│   ├── check_executor.go  # Check command execution
│   ├── cli.go         # Model CLI commands
│   ├── command_loader.go  # Custom command loading
│   ├── download.go    # HuggingFace model download
│   ├── model.go       # ONNX model loading & inference
│   └── prompt.go      # Prompt builders
├── tests/             # Test files
│   ├── data/          # Test data files
│   ├── servers/       # Mock servers (HTTP/gRPC/SSH)
│   └── bench_test.go  # Performance benchmarks
└── plans/             # Requirement & progress documents
```

## Running Tests

```bash
# All tests (skip slow LLM tests)
go test -count=1 ./internal/... ./pkg/...

# With benchmarks
go test -bench=. -benchtime=100ms ./tests/

# Mock servers for manual testing
go run ./tests/servers/http/   # port 18080
go run ./tests/servers/ssh/    # port 18082
go run ./tests/servers/grpc/   # port 18081
```

Test summary: 75/78 tests passing (96.2%), 3 intentional skips (SSH/Go 1.25, gRPC structpb).

## Open Source Dependencies

### Direct

| Package | Repository | License | Purpose |
|---------|-----------|---------|---------|
| onnxruntime-genai_purego | [getcharzp/onnxruntime-genai_purego](https://github.com/getcharzp/onnxruntime-genai_purego) | MIT | ONNX model inference via purego |
| tview | [rivo/tview](https://github.com/rivo/tview) | MIT | Terminal UI framework |
| tcell | [gdamore/tcell](https://github.com/gdamore/tcell) | Apache-2.0 | Terminal cell handling |
| toml | [BurntSushi/toml](https://github.com/BurntSushi/toml) | MIT | TOML configuration parsing |

### Indirect

| Package | Repository | License | Purpose |
|---------|-----------|---------|---------|
| purego | [ebitengine/purego](https://github.com/ebitengine/purego) | Apache-2.0 | CGo-free C function calls |
| golang.org/x/crypto | [go/x/crypto](https://cs.opensource.google/go/x/crypto) | BSD-3-Clause | SSH client |
| golang.org/x/net | [go/x/net](https://cs.opensource.google/go/x/net) | BSD-3-Clause | HTTP/gRPC networking |
| golang.org/x/sys | [go/x/sys](https://cs.opensource.google/go/x/sys) | BSD-3-Clause | System calls (Windows locale) |
| golang.org/x/term | [go/x/term](https://cs.opensource.google/go/x/term) | BSD-3-Clause | Terminal I/O |
| golang.org/x/text | [go/x/text](https://cs.opensource.google/go/x/text) | BSD-3-Clause | Text encoding/transform |
| google.golang.org/grpc | [grpc/grpc-go](https://github.com/grpc/grpc-go) | Apache-2.0 | gRPC client |
| google.golang.org/protobuf | [protocolbuffers/protobuf-go](https://github.com/protocolbuffers/protobuf-go) | BSD-3-Clause | Protocol Buffers |
| genproto | [googleapis/go-genproto](https://github.com/googleapis/go-genproto) | Apache-2.0 | gRPC status types |
| go-colorful | [lucasb-eyer/go-colorful](https://github.com/lucasb-eyer/go-colorful) | MIT | Color manipulation |
| go-runewidth | [mattn/go-runewidth](https://github.com/mattn/go-runewidth) | MIT | Rune width calculation |
| uniseg | [rivo/uniseg](https://github.com/rivo/uniseg) | MIT | Unicode segmentation |
| encoding | [gdamore/encoding](https://github.com/gdamore/encoding) | Apache-2.0 | Character encoding |
| gotool | [up-zero/gotool](https://github.com/up-zero/gotool) | MIT | Build tooling |

## License

MIT License - see [LICENSE](LICENSE) file for details.

Third-party components are distributed under their respective licenses as listed above.
