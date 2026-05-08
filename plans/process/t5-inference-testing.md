# T5: LLM Inference Testing

## 任务状态: 已完成

| 子任务 | 状态 | 说明 |
|--------|------|------|
| T5.0 Pre-check | 已完成 | 验证 data/prod-migration-form-uat.txt 存在 |
| T5.1 tests/llm_inference_test.go | 已完成 | DeepSeek R1, GLM5.1, ModelComparison 测试 |
| T5.2 tests/scenario_migration_test.go | 已完成 | FullPipeline, TUIInteraction, TaskCoverage 测试 |
| T5.3 runTestMode 修改 | 已完成 | 支持多模型对比 (DeepSeek + GLM5.1) |
| T5.4 运行测试验证 | 已完成 | 所有核心测试通过 |
| T5.4a 修复网络依赖加载 | 已完成 | DefaultGenAILibraryPath() 改为绝对路径 |

## 测试结果

### TestDeepSeekR1_MigrationParse
- **状态**: PASS (16.28s)
- Parser: 17 tasks
- LLM (DeepSeek R1): 14 tasks (82% coverage)
- 模型路径: models/deepseek-r1-distill-qwen-1.5B/cpu_and_mobile/cpu-int4-rtn-block-32-acc-level-4

### TestGLM51_MigrationParse
- **状态**: SKIP (模型未下载)
- 原因: models/glm5.1-distill-onnx 目录不存在
- 操作: `xsh model download glm5.1`

### TestModelComparison
- **状态**: PASS (14.48s)
- Baseline (Parser): 17 tasks
- deepseek-r1: 14 tasks (82% coverage)

### TestMigrationFullPipeline
- **状态**: PASS (88.81s)
- Parse: 17 tasks
- LLM Analyze: 14 tasks
- Execute: 14 tasks (8 OK, 6 FAIL 来自 HTTP 502 - mock server 未运行)
- 结果保存: tests/data/migration-pipeline-result.md

### TestMigrationTUIInteraction
- **状态**: PASS (64.78s)
- 加载 17 个任务
- 显示 9 个对话框 (5 ask + 4 check)
- TUI 流程正确: Parse → Analyze → Step-by-step execution

### TestMigrationTaskCoverage
- **状态**: PASS
- http: 6, ask: 5, check: 4, wait: 2
- 所有 4 种任务类型覆盖

## 修复内容

### DefaultGenAILibraryPath 路径修复 (T5.4a)
- **问题**: 相对路径 `lib/onnxruntime-genai.dll` 从 tests/ 目录读取时找不到，触发网络下载挂起
- **修复**: 使用 `runtime.Caller(0)` 从源文件位置计算项目根目录，返回绝对路径
- **影响**: pkg/llm/model.go DefaultGenAILibraryPath/DefaultOnnxRuntimeLibraryPath

## 输出文件
- tests/data/deepseek-r1-inference-test-result.md
- tests/data/model-comparison-result.md
- tests/data/migration-pipeline-result.md

## 未完成项
- GLM5.1 模型下载（需执行 `xsh model download glm5.1`）
- test servers (HTTP mock on port 18080) 未运行，导致 FullPipeline 中 6 个 HTTP 任务失败（非 LLM 相关）
