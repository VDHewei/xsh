# T7: TUI 自定义 Commands 支持 - 启动提示词

## 当前状态
- T6 待完成: 全面测试 (T7 前置依赖)
- TUI 已有基础框架: `internal/tui/tui.go` (任务列表、输入、确认对话框)
- Parser 已有任务解析: `internal/parser/parser.go`

## 任务目标
支持 `commands/` 目录下 `.md` 命令文件, 通过 LLM 解析 markdown 为任务列表, 集成到 TUI。

## 命令文件格式规范 (commands/xxx.md)
```markdown
# 命令名
## 描述
简要描述命令用途

## 任务
[GET] http://localhost:8080/health
[POST] http://localhost:8080/backup
@ask: 是否继续?
@check: 验证结果
@wait: 5min
```

## 数据模型 (需新增)

在 `internal/types/types.go` 或新文件 `pkg/llm/command_loader.go` 中:
```go
type CustomCommand struct {
    Name    string
    File    string
    Content string
    Tasks   []*types.Task
}

type CommandLoader struct {
    dir      string
    commands map[string]*CustomCommand
}
```

## 需要新增的文件

### 1. `pkg/llm/command_loader.go`
```go
func NewCommandLoader(dir string) *CommandLoader
func (l *CommandLoader) Scan() ([]string, error)            // 扫描 commands/ 目录
func (l *CommandLoader) Load(name string) (*CustomCommand, error)  // 加载 .md 文件
func (l *CommandLoader) ParseCommand(content string) (*CustomCommand, error)  // LLM 解析 markdown 为任务
```

**逻辑**: `Scan` 扫描目录 → `Load` 读取文件 → `ParseCommand` 用 LLM 解析 ## 任务 区块为 `[]*types.Task`

### 2. `commands/deploy.md` (示例命令文件)
创建示例命令文件用于测试

## 需要修改的文件

### 3. `internal/tui/tui.go`
- `setupUI()`: 增加 commands 选择区 (左侧列表或下拉)
- `loadCustomCommands()`: 加载自定义命令列表
- `runCustomCommand(name)`: 加载 → 显示任务列表 → 执行

## 验收标准
1. 放置 `commands/deploy.md` 后, TUI 中可选 "deploy" 命令
2. 选中后加载并显示任务列表
3. 支持执行/ask/check 交互
4. LLM 能在无模型时优雅降级 (使用 regex 正则解析替代)

## 参考代码
- TUI 框架: `internal/tui/tui.go` (tview 框架)
- 任务解析: `internal/parser/parser.go`
- 任务列表更新: `tui.go:126-140` (updateTaskList)
- 任务执行: `tui.go:142-186` (runNextTask)
