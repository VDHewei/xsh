# 生产部署命令

## 描述
自动化生产环境部署流程: 检查服务健康, 部署新版本, 等待就绪, 验证部署结果.

## 任务
[GET] http://localhost:18080/health
@ask: 确认开始部署?
[POST] http://localhost:18080/post
header: Content-Type=application/json
body: {"action":"deploy"}
@wait: 2min
@check: 验证部署结果
