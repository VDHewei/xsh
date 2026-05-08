# T8: 开源项目审核与 README - 完成

**完成时间**: 2026-05-08

## 子任务进度

| 子任务 | 状态 | 备注 |
|--------|------|------|
| T8.1 整理开源依赖清单 | 已完成 | 18 个依赖全部审核完毕 |
| T8.2 更新 README.md | 已完成 | 完整特性列表、依赖表、配置说明 |
| T8.3 更新 plans/llm.md | 已完成 | 审核状态全部标记为完成 |

## 许可证汇总

| 许可证 | 数量 | 项目 |
|--------|------|------|
| MIT | 8 | onnxruntime-genai_purego, tview, toml, go-colorful, go-runewidth, uniseg, gotool |
| Apache-2.0 | 5 | tcell, purego, grpc, encoding, genproto |
| BSD-3-Clause | 5 | golang.org/x/crypto, net, sys, term, text; google.golang.org/protobuf |

所有依赖均为宽松许可证 (MIT/Apache-2.0/BSD-3), 无 GPL/AGPL 依赖。

## README.md 更新内容

- 项目简介 (中英文双语)
- 功能特性 (Task parsing, HTTP full methods, TUI + config + i18n, LLM, CLI)
- 安装 (go install + build from source)
- 使用示例 (TUI, CLI, config, model management)
- 任务文件格式 + 自定义命令格式
- 项目结构 (当前完整状态)
- 测试说明 (75/78 passing, 96.2%)
- 开源依赖表 (18 个依赖, 含仓库/许可证/用途)
- License 声明
