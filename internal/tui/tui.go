package tui

import (
	"fmt"
	"os"
	"time"

	"github.com/VDHewei/xsh/internal/executor"
	"github.com/VDHewei/xsh/internal/parser"
	"github.com/VDHewei/xsh/internal/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// App TUI 应用
type App struct {
	app        *tview.Application
	pages      *tview.Pages
	taskList   *tview.TextView
	inputField *tview.InputField
	progress   *tview.TextView
	tasks      []*types.Task
	currentIdx int
	results    []string
	inputFile  string
	outputFile string
	exec       *executor.Executor
}

// NewApp 创建新应用
func NewApp() *App {
	return &App{
		app:        tview.NewApplication(),
		pages:      tview.NewPages(),
		taskList:   tview.NewTextView(),
		inputField: tview.NewInputField(),
		progress:   tview.NewTextView(),
		tasks:      nil,
		currentIdx: 0,
		results:    []string{},
		exec:       executor.NewExecutor(),
	}
}

// RunInteractive 运行交互模式
func RunInteractive() {
	app := NewApp()
	app.setupUI()
	if err := app.app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}

func (a *App) setupUI() {
	// 标题
	header := tview.NewTextView()
	header.SetText("xsh - 任务执行工具")
	header.SetTextColor(tcell.ColorGreen)
	header.SetTextAlign(tview.AlignCenter)

	// 任务列表
	a.taskList.SetBorder(true).SetTitle("任务列表")
	a.taskList.SetDynamicColors(true)

	// 进度显示
	a.progress.SetBorder(true).SetTitle("执行进度")
	a.progress.SetText("等待加载任务文件...")

	// 输入区域
	inputLabel := tview.NewTextView()
	inputLabel.SetText("加载任务文件: ")
	inputLabel.SetTextAlign(tview.AlignRight)

	a.inputField.SetBorder(true).SetTitle("输入文件路径")
	a.inputField.SetPlaceholder("请输入任务文件路径 (如: tests/data/prod-migration-form-uat.txt)")
	a.inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			a.loadTasks(a.inputField.GetText())
		}
	})

	// 主布局
	mainGrid := tview.NewGrid().
		SetRows(3, 0, 3, 3, 3).
		SetColumns(0, 0).
		AddItem(header, 0, 0, 1, 2, 0, 0, false).
		AddItem(a.taskList, 1, 0, 1, 2, 0, 0, false).
		AddItem(a.progress, 2, 0, 1, 2, 0, 0, false).
		AddItem(inputLabel, 3, 0, 1, 1, 0, 0, false).
		AddItem(a.inputField, 3, 1, 1, 1, 0, 0, false)

	a.pages.AddPage("main", mainGrid, true, true)
	a.app.SetRoot(a.pages, true)
}

func (a *App) loadTasks(filename string) {
	if filename == "" {
		a.progress.SetText("[red]请输入文件路径[red]")
		return
	}

	tasks, err := parser.ParseFile(filename)
	if err != nil {
		a.progress.SetText(fmt.Sprintf("[red]解析文件失败: %v[red]", err))
		return
	}

	a.tasks = tasks
	a.currentIdx = 0
	a.inputFile = filename

	// 显示任务列表
	a.updateTaskList()
	a.progress.SetText(fmt.Sprintf("已加载 %d 个任务，按 Enter 开始执行", len(tasks)))

	// 设置按键处理
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			a.runNextTask()
		}
		return event
	})
}

func (a *App) updateTaskList() {
	var text string
	for i, task := range a.tasks {
		prefix := "  "
		if i == a.currentIdx {
			prefix = "> "
		}
		completed := ""
		if i < a.currentIdx {
			completed = " [green]✓[green]"
		}
		text += fmt.Sprintf("%s[%d] %s%s\n", prefix, i+1, task.Raw, completed)
	}
	a.taskList.SetText(text)
}

func (a *App) runNextTask() {
	if a.currentIdx >= len(a.tasks) {
		a.progress.SetText("[green]所有任务执行完成![green]")
		a.saveResults()
		return
	}

	task := a.tasks[a.currentIdx]
	a.progress.SetText(fmt.Sprintf("正在执行 [%d/%d]: %s", a.currentIdx+1, len(a.tasks), task.Raw))

	// 根据任务类型执行
	result := a.executeTask(task)
	a.results = append(a.results, result)

	// 显示结果
	switch task.Type {
	case types.TaskTypeAsk:
		// 询问用户
		a.showConfirmDialog(task.Ask.Prompt)
		return
	case types.TaskTypeCheck:
		// 检查并询问
		a.showConfirmDialog(task.Check.Prompt)
		return
	case types.TaskTypeWait:
		// 等待后继续
		duration, _ := time.ParseDuration(task.Wait.Duration)
		time.AfterFunc(duration, func() {
			a.app.QueueUpdate(func() {
				a.currentIdx++
				a.updateTaskList()
				a.runNextTask()
			})
		})
		return
	}

	a.currentIdx++
	a.updateTaskList()

	// 继续执行下一个任务
	a.app.QueueUpdate(func() {
		a.runNextTask()
	})
}

func (a *App) executeTask(task *types.Task) string {
	result := a.exec.ExecuteTasks([]*types.Task{task})
	if len(result) > 0 {
		r := result[0]
		if r.Success {
			return r.Output
		}
		return fmt.Sprintf("[ERROR] %s: %v", r.Output, r.Error)
	}
	return fmt.Sprintf("[SKIP] %s", task.Raw)
}

func (a *App) showConfirmDialog(prompt string) {
	modal := tview.NewModal()
	modal.SetText(prompt)
	modal.AddButtons([]string{"继续", "跳过", "退出"})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		a.pages.HidePage("confirm")
		switch buttonLabel {
		case "继续":
			a.currentIdx++
			a.updateTaskList()
			a.runNextTask()
		case "跳过":
			a.currentIdx++
			a.updateTaskList()
			a.runNextTask()
		case "退出":
			a.app.Stop()
		}
	})
	a.pages.AddPage("confirm", modal, true, true)
}

func (a *App) saveResults() {
	if a.outputFile == "" {
		return
	}
	// TODO: 保存结果到文件
}

