# 数据迁移命令

## 描述
执行数据库迁移: 备份现有数据, 执行数据迁移脚本, 验证迁移完整性.

## 任务
[GET] http://localhost:18080/health
@ask: 确认开始数据迁移?
[POST] http://localhost:18080/post
header: Content-Type=application/json
body: {"action":"backup","target":"database"}
@wait: 3min
[PUT] http://localhost:18080/put
body: {"action":"migrate","from":"v1","to":"v2"}
@wait: 1min
[GET] http://localhost:18080/health
@check: 验证数据迁移完整性, 服务是否正常?
