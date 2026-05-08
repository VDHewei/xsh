# T3: SSH/gRPC 执行器 + 重试/超时 - 进度 ✅

| 子任务 | 状态 | 备注 |
|--------|------|------|
| T3.1 executeSSH() | 已完成 | golang.org/x/crypto/ssh, 支持命令执行 |
| T3.2 executeGRPC() | 已完成 | google.golang.org/grpc, 通用方法调用 |
| T3.3 重试机制 | 已完成 | RetryConfig, 3次重试, 2x指数退避 |
| T3.4 超时控制 | 已完成 | HTTP 30s, gRPC 30s context, SSH 10s |
| T3.5 错误分类返回 | 已完成 | HTTP_TIMEOUT/CONNECTION/DNS/TLS/ERROR |

## 测试结果
- 通过: 20/20
- 失败: 0
- 通过原因: 7种HTTP方法(含header/body), 失败/重试/连接拒绝/错误分类全部验证通过

## 新增文件
- `internal/executor/retry.go` - RetryConfig, Do()
- `internal/executor/executor.go` - Executor struct, executeHTTP/SSH/gRPC/Wait

## 修改文件
- `internal/executor/executor_test.go` - 20个测试用例
- `internal/tui/tui.go` - executeTask委托给executor, 移除mock函数
