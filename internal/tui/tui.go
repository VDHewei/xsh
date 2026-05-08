package tui

import (
	"fmt"
	"os"
	"time"

	"github.com/VDHewei/xsh/internal/config"
	"github.com/VDHewei/xsh/internal/executor"
	"github.com/VDHewei/xsh/internal/i18n"
	"github.com/VDHewei/xsh/internal/parser"
	"github.com/VDHewei/xsh/internal/types"
	llm "github.com/VDHewei/xsh/pkg/llm"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// App TUI 应用
type App struct {
	app           *tview.Application
	pages         *tview.Pages
	taskList      *tview.TextView
	inputField    *tview.InputField
	progress      *tview.TextView
	tasks         []*types.Task
	currentIdx    int
	results       []string
	inputFile     string
	outputFile    string
	exec          *executor.Executor
	model         *llm.Model
	config        *config.Config
	i18n          *i18n.Manager
	commandLoader *llm.CommandLoaderImpl
	commandList   *tview.List
	commands      []string
}

// NewApp 创建新应用
func NewApp(cfg *config.Config) *App {
	mgr := i18n.New(cfg.Language, cfg.I18N)
	return &App{
		app:           tview.NewApplication(),
		pages:         tview.NewPages(),
		taskList:      tview.NewTextView(),
		inputField:    tview.NewInputField(),
		progress:      tview.NewTextView(),
		tasks:         nil,
		currentIdx:    0,
		results:       []string{},
		exec:          executor.NewExecutor(),
		config:        cfg,
		i18n:          mgr,
		commandLoader: llm.NewCommandLoader("commands"),
	}
}

// SetLLMModel 注入 LLM 模型
func (a *App) SetLLMModel(model *llm.Model) {
	a.model = model
	a.exec.SetLLMModel(model)
}

// RunInteractive 运行交互模式
func RunInteractive() {
	cfg := config.Load()
	app := NewApp(cfg)
	app.setupUI()
	if err := app.app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}

func (a *App) setupUI() {
	// 标题
	header := tview.NewTextView()
	header.SetText(a.i18n.Header())
	header.SetTextColor(parseColor(a.config.Style.Header))
	header.SetTextAlign(tview.AlignCenter)

	// 命令列表
	a.commandList = tview.NewList()
	a.commandList.SetBorder(true).SetTitle(a.i18n.CmdListTitle())
	a.commandList.SetSelectedFunc(func(idx int, mainText, secondaryText string, shortcut rune) {
		a.loadCustomCommand(mainText)
		a.app.SetFocus(a.inputField)
	})

	// 任务列表
	a.taskList.SetBorder(true).SetTitle(a.i18n.TaskListTitle())
	a.taskList.SetDynamicColors(true)

	// 进度显示
	a.progress.SetBorder(true).SetTitle(a.i18n.ProgressTitle())
	a.progress.SetText(a.i18n.WaitingText())

	// 输入区域
	inputLabel := tview.NewTextView()
	inputLabel.SetText(a.i18n.InputLabel())
	inputLabel.SetTextAlign(tview.AlignRight)

	a.inputField.SetBorder(true).SetTitle(a.i18n.InputTitle())
	a.inputField.SetPlaceholder(a.i18n.InputPlaceholder())
	a.inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			a.loadTasks(a.inputField.GetText())
		}
	})

	// 加载命令列表
	a.scanCommands()

	// 主布局: 两列 Grid
	cmdWidth := a.config.Layout.CommandListWidth
	if cmdWidth <= 0 {
		cmdWidth = 25
	}
	mainGrid := tview.NewGrid().
		SetRows(3, 0, 3, 3, 3).
		SetColumns(cmdWidth, 0).
		AddItem(header, 0, 0, 1, 2, 0, 0, false).
		AddItem(a.commandList, 1, 0, 4, 1, 0, 0, false).
		AddItem(a.taskList, 1, 1, 1, 1, 0, 0, false).
		AddItem(a.progress, 2, 1, 1, 1, 0, 0, false).
		AddItem(inputLabel, 3, 1, 1, 1, 0, 0, false).
		AddItem(a.inputField, 4, 1, 1, 1, 0, 0, false)

	a.pages.AddPage("main", mainGrid, true, true)
	a.app.SetRoot(a.pages, true)

	// 全局按键: Tab 切换焦点, Enter 执行任务
	a.setupKeyBindings()
}

func (a *App) setupKeyBindings() {
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			// 在 inputField 和 commandList 之间切换焦点
			if a.inputField.HasFocus() {
				a.app.SetFocus(a.commandList)
			} else {
				a.app.SetFocus(a.inputField)
			}
			return nil
		case tcell.KeyEnter:
			// Enter 执行下一个任务
			if a.tasks != nil && len(a.tasks) > 0 {
				a.runNextTask()
				return nil
			}
		}
		return event
	})
}

func (a *App) scanCommands() {
	a.commandList.Clear()
	a.commands = nil

	names, err := a.commandLoader.Scan()
	if err != nil {
		return
	}

	for _, name := range names {
		a.commands = append(a.commands, name)
		a.commandList.AddItem(name, "", 0, nil)
	}
}

func (a *App) loadCustomCommand(name string) {
	cmd, err := a.commandLoader.Load(name)
	if err != nil {
		a.progress.SetText(a.errTag(a.i18n.ParseFailedFmt(err)))
		return
	}

	a.tasks = cmd.Tasks
	a.currentIdx = 0
	a.results = nil
	a.inputFile = fmt.Sprintf("command:%s", name)

	a.updateTaskList()
	a.progress.SetText(a.i18n.TasksLoadedCmdFmt(name, len(cmd.Tasks)))
}

func (a *App) loadTasks(filename string) {
	if filename == "" {
		a.progress.SetText(a.errTag(a.i18n.InvalidPathText()))
		return
	}

	tasks, err := parser.ParseFile(filename)
	if err != nil {
		a.progress.SetText(a.errTag(a.i18n.ParseFailedFmt(err)))
		return
	}

	a.tasks = tasks
	a.currentIdx = 0
	a.inputFile = filename
	a.results = nil

	a.updateTaskList()
	a.progress.SetText(a.i18n.TasksLoadedFmt(len(tasks)))
}

func (a *App) updateTaskList() {
	okColor := a.config.Style.Success
	var text string
	for i, task := range a.tasks {
		prefix := "  "
		if i == a.currentIdx {
			prefix = "> "
		}
		completed := ""
		if i < a.currentIdx {
			completed = fmt.Sprintf(" [%s]✓[%s]", okColor, okColor)
		}
		text += fmt.Sprintf("%s[%d] %s%s\n", prefix, i+1, task.Raw, completed)
	}
	a.taskList.SetText(text)
}

func (a *App) runNextTask() {
	if a.currentIdx >= len(a.tasks) {
		a.progress.SetText(a.okTag(a.i18n.AllDoneText()))
		a.saveResults()
		return
	}

	task := a.tasks[a.currentIdx]
	a.progress.SetText(a.i18n.TaskExecutingFmt(a.currentIdx+1, len(a.tasks), task.Raw))

	result := a.executeTask(task)
	a.results = append(a.results, result)

	switch task.Type {
	case types.TaskTypeAsk:
		a.showAskResultDialog(result)
		return
	case types.TaskTypeCheck:
		a.showCheckResultDialog(result)
		return
	case types.TaskTypeWait:
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

// showAskResultDialog 展示 LLM Ask 推理结果
func (a *App) showAskResultDialog(result string) {
	modal := tview.NewModal()
	displayText := result
	if len(displayText) > 200 {
		displayText = displayText[:200] + "..."
	}
	modal.SetText(fmt.Sprintf("[%s]\n\n%s", a.i18n.AskTitle(), displayText))
	modal.AddButtons([]string{
		a.i18n.ContinueButton(),
		a.i18n.SkipButton(),
		a.i18n.ExitButton(),
	})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		a.pages.HidePage("askResult")
		switch buttonLabel {
		case a.i18n.ContinueButton(), a.i18n.SkipButton():
			a.currentIdx++
			a.updateTaskList()
			a.runNextTask()
		case a.i18n.ExitButton():
			a.app.Stop()
		}
	})
	a.pages.AddPage("askResult", modal, true, true)
}

// showCheckResultDialog 展示 LLM Check 检查结果
func (a *App) showCheckResultDialog(result string) {
	modal := tview.NewModal()
	modal.SetText(fmt.Sprintf("[%s]\n\n%s", a.i18n.CheckTitle(), result))
	modal.AddButtons([]string{
		a.i18n.ContinueButton(),
		a.i18n.SkipButton(),
		a.i18n.ExitButton(),
	})
	modal.SetDoneFunc(func(buttonIndex int, buttonLabel string) {
		a.pages.HidePage("checkResult")
		switch buttonLabel {
		case a.i18n.ContinueButton(), a.i18n.SkipButton():
			a.currentIdx++
			a.updateTaskList()
			a.runNextTask()
		case a.i18n.ExitButton():
			a.app.Stop()
		}
	})
	a.pages.AddPage("checkResult", modal, true, true)
}

func (a *App) saveResults() {
	if a.outputFile == "" {
		return
	}
	// TODO: 保存结果到文件
}

// --- helpers ---

func (a *App) errTag(text string) string {
	return fmt.Sprintf("[%s]%s[%s]", a.config.Style.Error, text, a.config.Style.Error)
}

func (a *App) okTag(text string) string {
	return fmt.Sprintf("[%s]%s[%s]", a.config.Style.Success, text, a.config.Style.Success)
}

func parseColor(name string) tcell.Color {
	switch name {
	case "red":
		return tcell.ColorRed
	case "green":
		return tcell.ColorGreen
	case "yellow":
		return tcell.ColorYellow
	case "blue":
		return tcell.ColorBlue
	case "magenta":
		return tcell.ColorPurple
	case "cyan":
		return tcell.ColorLightCyan
	case "white":
		return tcell.ColorWhite
	case "orange":
		return tcell.ColorOrange
	default:
		return tcell.ColorWhite
	}
}
