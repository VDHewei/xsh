# T1: Model CLI 接入 main.go 路由 - 进度 ✅

| 子任务 | 状态 | 备注 |
|--------|------|------|
| T1.1 main.go model 子命令路由 | 已完成 | model search/list/select/download 四命令 |
| T1.2 GLM5.1 候选列表 | 已完成 | DefaultCandidateModels() 含 deepseek + glm5.1 |
| T1.3 download --dir 参数 | 已完成 | handleModelDownload 支持 --dir |
| T1.4 代理/镜像下载 | 已完成 | DownloadWithProxy/DownloadWithMirror |

## 测试结果
- 通过: 4/4 功能入口路由全部实现
- 失败: /
- 通过原因: main.go model路由, cli.go ParseModelCommand四命令, download.go resolveModelRepo候选映射, model.go DefaultCandidateModels 全部实现

## 注意
- `ModelCandidate` 结构体未单独定义，改用 `map[string]string` 直接映射，功能等效
- GLM5.1 模型尚未实际下载到本地 models/ 目录
