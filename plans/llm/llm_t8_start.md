# T8: 开源项目审核与 README - 启动提示词

## 当前状态
所有功能模块 (T1-T7) 已实现或规划完成, 进入收尾阶段。

## 任务目标
整理开源依赖清单, 写入 README.md, 更新审核状态。

## 需要修改的文件

### 1. `README.md`
需要整合的内容:
- 项目简介: xsh - 任务执行工具 (HTTP/SSH/gRPC + LLM 推理)
- 功能特性列表
- 安装和使用说明
- 开源依赖清单 (需要标注许可证)

### 2. `plans/llm.md`
更新开源项目清单审核状态

## 开源依赖清单 (待审核)

| 项目 | 仓库 | 许可证 | 用途 | 审核 |
|------|------|--------|------|------|
| onnxruntime-genai_purego | github.com/getcharzp/onnxruntime-genai_purego | - | ONNX 模型推理 | 待审核 |
| tview | github.com/rivo/tview | MIT | 终端 UI | 待审核 |
| tcell | github.com/gdamore/tcell | Apache-2.0 | 终端单元格 | 待审核 |
| purego | github.com/ebitengine/purego | Apache-2.0 | CGo-free FFI | 待审核 |
| golang.org/x/crypto | cs.opensource.google/go/x/crypto | BSD-3 | SSH 客户端 | 待审核 |
| google.golang.org/grpc | github.com/grpc/grpc-go | Apache-2.0 | gRPC 客户端 | 待审核 |
| google.golang.org/protobuf | github.com/protocolbuffers/protobuf-go | BSD-3 | protobuf | 待审核 |

## 执行步骤
1. **Step 1**: 读取 `go.mod` 获取完整依赖列表
2. **Step 2**: 确认每个依赖的许可证 (检查仓库 LICENSE 文件)
3. **Step 3**: 更新 `plans/llm.md` 审核状态
4. **Step 4**: 编写 README.md 包含:
   - 项目简介
   - 安装: `go install github.com/VDHewei/xsh/cmd/xsh@latest`
   - 使用示例
   - 开源依赖表
   - 许可证声明

## 验收标准
1. README.md 包含完整开源依赖清单
2. 仓库地址、用途、许可证明确
3. `plans/llm.md` 审核状态已标记

## 参考
- go.mod: 项目根目录
- 现有开源清单: `plans/llm.md:66-77`
