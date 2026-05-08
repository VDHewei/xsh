# T6: 全面测试覆盖 - 启动提示词

## 当前状态
- T3 已完成: Executor SSH/gRPC/HTTP + 重试/超时
- T5 待完成: LLM 推理测试 (T6 前置依赖)
- `cmd/xsh/onnx_test.go` 尚未归档

## 任务目标
全面补充测试覆盖: 单元测试 + 场景链路 + 非代码测试 + 性能测试 + 代码归档。

## 需要新增的文件

### 1. `internal/parser/parser_test.go` (补充)
```go
func TestParseHTTPWithHeaders(t *testing.T)     // 解析带 header 的 HTTP 行
func TestParseHTTPWithBody(t *testing.T)        // 解析带 body 的 HTTP 行
func TestParseGRPCTask(t *testing.T)            // 解析 gRPC 任务行
func TestParseSSHTask(t *testing.T)             // 解析 SSH 任务行
func TestParseMixedTasks(t *testing.T)          // 混合任务行解析 (模糊测试)
```

### 2. `internal/executor/executor_test.go` (补充)
HTTP 7 种方法全覆盖:
```go
func TestExecuteHTTP_GET/POST/PUT/PATCH/DELETE/HEAD/OPTIONS(t *testing.T)
func TestExecuteHTTP_WithHeaders(t *testing.T)   // 自定义 header
func TestExecuteHTTP_Timeout(t *testing.T)       // 超时场景
func TestExecuteHTTP_Retry(t *testing.T)         // 重试场景
func TestExecuteHTTP_Failure(t *testing.T)       // 4xx/5xx
func TestExecuteSSH_Success/Failure(t *testing.T)
func TestExecuteGRPC_Success/Failure(t *testing.T)
func TestRetryConfig_Do(t *testing.T)            // 重试逻辑
func TestClassifyError(t *testing.T)             // 错误分类
```

### 3. `tests/bench_test.go` (新增)
```go
func BenchmarkLLMInference(b *testing.B)        // LLM 推理延迟
func BenchmarkStreamInference(b *testing.B)      // 流式推理
func BenchmarkHTTPExecution(b *testing.B)        // HTTP 执行
```

### 4. `tests/manual_test.md` (新增)
- Agent 随机测试: 随机 URL/method/header 组合
- 人机交互测试: TUI 完整流程操作记录

### 5. 测试分类
| 分类 | 测试 |
|------|------|
| 通过测试 | TestExecuteHTTP_GET, TestExecuteHTTP_POST, TestParseFile, TestExecuteSSH_Success |
| 非通过测试 | TestExecuteHTTP_Timeout, TestExecuteHTTP_Failure, TestExecuteSSH_Failure |
| 模糊测试 | TestParseMixedTasks, TestExecuteMultipleTasks, TestRetryConfig_Do |

## 需要修改/归档的文件

### 6. `cmd/xsh/onnx_test.go` → 归档到 `tests/cmd_onnx_test.go`
- 移动文件
- `buildTaskPrompt` 等辅助函数统一到 `pkg/llm/`

### 7. `models/deepseek-r1-distill-qwen-1.5B` 路径
- 确认模型路径在测试中正确使用

## 验收标准
1. `go test ./...` 全部通过
2. 7 种 HTTP 方法全部有测试覆盖
3. 错误场景 (失败/重试/超时/成功/异常) 全覆盖
4. 测试分类: 通过/非通过/模糊 各至少一组
5. `cmd/xsh/` 仅有 `main.go` (onnx_test.go 已归档)
6. 性能基准记录到 `tests/data/bench_results.md`

## 参考代码
- 现有测试: `internal/executor/executor_test.go` (20 个已有测试)
- 现有测试: `internal/parser/parser_test.go`
- Mock 服务器: `tests/servers/http/main.go` (启动方式: `go run tests/servers/http/main.go`)
- `cmd/xsh/onnx_test.go` 待归档
