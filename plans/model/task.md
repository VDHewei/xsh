# T1 Entry Layer & GLM5.1 Support - 任务清单

## 状态: ✅ 已完成

| # | 任务 | 状态 | 文件 |
|---|------|------|------|
| 1 | 定义 ModelCandidate 结构体 | ✅ 已完成 | `internal/types/types.go` |
| 2 | 修复 DefaultCandidateModels() 返回类型 | ✅ 已完成 | `pkg/llm/model.go`, `pkg/llm/download.go` |
| 3 | 增强 GetModelPath 候选名匹配 | ✅ 已完成 | `pkg/llm/model.go` |
| 4 | 增强 xsh model list 候选标记 | ✅ 已完成 | `pkg/llm/model.go`, `pkg/llm/cli.go` |
| 5 | 修复动态库重复下载 | ✅ 已完成 | `pkg/llm/model.go` |
| 6 | 修复 runTestMode 动态库调用 | ✅ 已完成 | `cmd/xsh/main.go` |
| 7 | GetModelPath 候选名测试 | ✅ 已完成 | `pkg/llm/llm_test.go` |
| 8 | model list 候选标记测试 | ✅ 已完成 | `pkg/llm/llm_test.go` |
| 9 | 动态库缓存测试 | ✅ 已完成 | `pkg/llm/llm_test.go` |
| 10 | 运行全部测试 | ✅ 已完成 | - |
| 11 | 更新 README (中英双版) | ✅ 已完成 | `README.md`, `README_CN.md` |

## 依赖关系
```
1 → 2 → 3 → 4 → 5/6 (并行) → 7/8/9 (并行测试) → 10 → 11
```

## 记录时间
2026-05-09 开始 → 2026-05-09 完成

## 测试结果
全部通过 (go test ./... -count=1): internal/config, internal/executor, internal/i18n, internal/parser, internal/tui, pkg/llm, tests

## 实现进度记录
详见 [plans/process/2026-05-09_t1_glm5.1_implementation.md](../process/2026-05-09_t1_glm5.1_implementation.md)
