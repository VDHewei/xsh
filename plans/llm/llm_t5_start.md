# T5: LLM 推理测试 (迁移文件识别) - 启动提示词

## 当前状态
- T1 已完成: 模型 CLI 就绪, DeepSeek R1 已下载到 `models/deepseek-r1-distill-qwen-1.5B/`
- T4 待完成: @ask/@check LLM 集成 (T5 前置依赖)
- GLM5.1 模型尚未下载 (`yasserrmd/glm5.1-distill-onnx`)

## 任务目标
验证 LLM 本地推理在迁移文件识别场景下的准确度和完整链路。

## 需要新增的文件

### 1. `tests/llm_inference_test.go`
```go
// TestDeepSeekR1_MigrationParse DeepSeek 解析迁移文件准确度
func TestDeepSeekR1_MigrationParse(t *testing.T)

// TestGLM51_MigrationParse GLM5.1 解析迁移文件准确度 (需先下载模型)
func TestGLM51_MigrationParse(t *testing.T)

// TestModelComparison 多模型对比测试
func TestModelComparison(t *testing.T)
```

**测试数据**: `tests/data/prod-migration-form-uat.txt` (已有)

**测试逻辑**:
- 用 `TaskAnalyzer.AnalyzeFile()` 解析迁移文件
- 验证 LLM 返回结果能正确识别迁移步骤 (预期 8 个步骤)
- 对比 DeepSeek R1 vs GLM5.1 的解析准确度
- 记录到 `tests/data/` 目录

### 2. `tests/scenario_migration_test.go`
```go
// TestMigrationFullPipeline parse → LLM分析 → execute → 结果验证 完整链路
func TestMigrationFullPipeline(t *testing.T)

// TestMigrationTUIInteraction TUI 完整交互流程
func TestMigrationTUIInteraction(t *testing.T)
```

**完整链路**:
1. `parser.ParseFile()` 解析任务文件
2. `TaskAnalyzer.AnalyzeContent()` LLM 分析
3. `executor.ExecuteTasks()` 执行
4. 验证结果正确性

### 3. `cmd/xsh/main.go` 修改
- `runTestMode()` (line 82) 增加 GLM5.1 模型测试对比分支
- 下载 GLM5.1 模型: `xsh model download glm5.1`

## 执行顺序
1. **前置**: 确认 T4 已完成 (ask_executor, check_executor 就绪)
2. **Step 1**: 下载 GLM5.1 模型 `xsh model download glm5.1`
3. **Step 2**: 编写 `tests/llm_inference_test.go`
4. **Step 3**: 编写 `tests/scenario_migration_test.go`
5. **Step 4**: 修改 `runTestMode()` 增加对比测试
6. **Step 5**: 运行测试: `go test ./tests/... -v`
7. **Loop**: 失败则分析原因修复, 成功才跳出

## 验收标准
1. DeepSeek R1 正确识别 8 个迁移步骤
2. GLM5.1 解析结果记录对比
3. 完整链路 parse → analyze → execute → verify 通过
4. 测试结果写入 `tests/data/` 目录

## 参考代码
- 任务分析器: `pkg/llm/analyzer.go` (AnalyzeFile, AnalyzeContent)
- 测试数据: `tests/data/prod-migration-form-uat.txt`
- 现有测试结果: `tests/data/deepseek-r1-inference-test-result.md`
- 模型下载: `pkg/llm/download.go` (DownloadFromHuggingFace)
- CLI 入口: `cmd/xsh/main.go:82-117` (runTestMode)
