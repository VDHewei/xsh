package llm

import "fmt"

// BuildTaskPrompt 构建任务分析提示
func BuildTaskPrompt(taskContent string) string {
	return fmt.Sprintf(`你是一个任务规划和执行助手。请分析以下迁移步骤，提取出结构化任务列表。

迁移步骤:
%s

请按以下格式输出，每行一个任务：
[GET] <url> 用于 GET 请求
header: Xxxx=xxx
[POST] <url> 用于 POST 请求
header: XXX=xxx
body: {}
@ask: <描述> 用于需要用户确认的步骤
@wait: <时长> 用于等待步骤 (如 @wait: 10min)
@check: <描述> 用于验证步骤

只输出任务列表，不要其他内容。`, taskContent)
}
