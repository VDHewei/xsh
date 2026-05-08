# T2: Mock 测试服务器 - 进度 ✅

| 子任务 | 状态 | 备注 |
|--------|------|------|
| T2.1 tests/servers/http | 已完成 | 7种方法 + echo/retry/timeout/fail 全部通过 |
| T2.2 tests/servers/grpc | 已完成 | 编译通过, grpcurl 不可用, 待 T3 中 Go client 测试 |
| T2.3 tests/servers/ssh | 已完成 | 编译通过, ssh client 连接受限环境, 待 T3 中 Go client 测试 |

## 测试结果
- 通过: HTTP GET/POST/PUT/PATCH/DELETE/HEAD/OPTIONS, retry(3次), fail, echo
- 失败: /
- 通过原因: HTTP 所有方法/场景 curl 验证通过; gRPC/SSH 编译通过, 待 T3 executor 验证

## 新增依赖
- google.golang.org/grpc v1.81.0
- google.golang.org/protobuf v1.36.11
- golang.org/x/crypto v0.50.0
