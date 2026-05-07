# xsh

<p align="center">
  <img src="docs/logo.svg" alt="xsh logo" width="128" height="128">
</p>

<p align="center">
  <strong>xsh</strong> - eXecute SHell tasks | 任务执行工具
</p>

<p align="center">
  <a href="README_CN.md">中文文档</a> | English
</p>

---

xsh is a Go-based task execution tool that reads task files (txt/md) and automates HTTP, SSH, gRPC task generation and execution. It features a TUI interactive mode, LLM-powered task analysis, and a structured command system.

## Features

- **Task File Parsing** - Read `.txt`/`.md` files and parse structured task commands
- **Command Syntax** - `@{command}:{description}` style command parsing
  - `@ask:` - Interactive user confirmation
  - `@wait:` - Timed wait with duration
  - `@check:` - Print task results and ask user to continue
- **HTTP Task Execution** - `[GET]`/`[POST]`/`[PUT]`/`[DELETE]` URL format
- **TUI Interactive Mode** - Terminal UI with task list, progress, and confirmation dialogs
- **LLM Integration** - Local ONNX GenAI model inference for task analysis and planning
  - Real inference via `onnxruntime-genai` (tokenization, autoregressive generation, KV cache, sampling)
  - Streaming output support (`-stream` flag)
  - Automatic fallback to mock inference when no model is available
- **Model Management** - `model search`, `model list`, `model select` CLI commands
- **CLI Mode** - `-i` input file, `-o` output results file

## Installation

```bash
go build -o xsh ./cmd/xsh
```

## Usage

### Interactive TUI Mode

```bash
./xsh
```

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

### Model Management

```bash
# Search models on HuggingFace
./xsh model search onnx

# List local models
./xsh model list

# Select a model
./xsh model select deepseek-r1-distill
```

### Test Mode

```bash
# Run ONNX GenAI test with mock inference (no model downloaded)
./xsh -test

# Run ONNX GenAI test with streaming output
./xsh -test -stream
```

### Prerequisites for LLM Inference

To use real LLM inference, you need:

1. **GenAI shared libraries** - Auto-downloaded to `lib/` on first use:
   - `onnxruntime-genai.dll` (Windows) / `.so` (Linux) / `.dylib` (macOS)
   - `onnxruntime.dll` (Windows) / `.so` (Linux) / `.dylib` (macOS)

2. **ONNX model directory** - Download from HuggingFace, e.g.:
   ```
   models/deepseek-r1-distill-qwen-1.5B/
   ├── genai_config.json
   ├── model.onnx
   ├── model.onnx.data
   ├── tokenizer.json
   └── tokenizer_config.json
   ```

## Task File Format

```text
# This is a comment
> This is also a comment

# HTTP requests
[GET] http://example.com/api/health
[POST] http://example.com/api/deploy

# Interactive commands
@ask: Confirm to proceed with deployment?
@wait: 10min
@check: Verify deployment status
```

## Project Structure

```
xsh/
├── cmd/xsh/           # Entry point and CLI
├── internal/
│   ├── executor/      # Task execution engine
│   ├── parser/        # Task file parser
│   ├── tui/           # Terminal UI
│   └── types/         # Type definitions
├── pkg/llm/           # LLM integration package
│   ├── analyzer.go    # Task analyzer
│   ├── cli.go         # Model CLI commands
│   ├── download.go    # HuggingFace model download
│   └── model.go       # ONNX model loading & inference
├── tests/data/        # Test data files
└── plans/             # Requirement documents
```

## Running Tests

```bash
go test ./... -v
```

## License

MIT
