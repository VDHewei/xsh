# Manual Testing Guide

This document describes manual testing procedures for the `xsh` CLI tool.
Run these tests when code changes affect core execution flow or user-facing behavior.

---

## 1. Prerequisites

### 1.1 Build the binary

```bash
cd d:/workspaces/go-src/xsh
go build -o xsh.exe ./cmd/xsh/
```

Verify the binary works:

```bash
./xsh.exe -h
```

### 1.2 Start mock servers

Start all three mock servers in separate terminals:

```bash
# Terminal 1: HTTP mock server (port 18080)
go run ./tests/servers/http/

# Terminal 2: SSH mock server (port 18082)
go run ./tests/servers/ssh/

# Terminal 3: gRPC mock server (port 18081)
go run ./tests/servers/grpc/
```

Verify servers are running:

```bash
curl http://localhost:18080/health          # {"status":"healthy"}
curl http://localhost:18081/health          # (if HTTP handler available)
# For SSH: ssh testuser@localhost -p 18082  # accepts any password
```

---

## 2. Task File Parsing

### 2.1 Basic HTTP tasks

Create `test_basic.txt`:

```
# Migration Plan
[GET] http://localhost:18080/health
[POST] http://localhost:18080/post
[PUT] http://localhost:18080/put
[PATCH] http://localhost:18080/patch
[DELETE] http://localhost:18080/delete
[HEAD] http://localhost:18080/head
[OPTIONS] http://localhost:18080/options
```

Run:

```bash
./xsh.exe -i test_basic.txt
```

**Expected**: 7 tasks executed, all showing success status codes. GET returns 200, POST returns 201.

### 2.2 Mixed task types

Create `test_mixed.txt`:

```
[GET] http://localhost:18080/health
@wait: 5min
@ask: Continue?
```

Run:

```bash
./xsh.exe -i test_mixed.txt
```

**Expected**: GET executes, Wait shows "5m0s" duration, Ask shows mock suggestion.

### 2.3 Invalid format handling

Create `test_invalid.txt`:

```
[MISSING_BRACKET http://example.com
@invalid: command
plain text line
```

Run:

```bash
./xsh.exe -i test_invalid.txt
```

**Expected**: Invalid lines are silently skipped, output may be empty or show 0 tasks.

---

## 3. HTTP Execution Scenarios

### 3.1 Retry on failure

Create `test_retry.txt`:

```
[GET] http://localhost:18080/retry
```

Run:

```bash
./xsh.exe -i test_retry.txt
```

**Expected**: Should show "success after retry" after automatic retries.

### 3.2 Permanent failure

Create `test_fail.txt`:

```
[GET] http://localhost:18080/fail
```

Run:

```bash
./xsh.exe -i test_fail.txt
```

**Expected**: Should show "500" error with "all retries exhausted".

### 3.3 Slow endpoint

Create `test_slow.txt`:

```
[GET] http://localhost:18080/slow?delay=2s
```

Run:

```bash
./xsh.exe -i test_slow.txt
```

**Expected**: Should complete successfully after 2-second delay, output contains "slow".

### 3.4 Connection error

Create `test_no_conn.txt`:

```
[GET] http://localhost:19999/nonexist
```

Run:

```bash
./xsh.exe -i test_no_conn.txt
```

**Expected**: Should show "HTTP_CONNECTION" error after retries.

---

## 4. SSH Execution

### 4.1 Basic SSH command

Create `test_ssh.txt`:

```
@ssh: echo hello world
```

> **Note**: Requires SSH server running on localhost:18082. The test user and password
> must be pre-configured. On Go 1.25, SSH handshake may stall due to `golang.org/x/crypto`
> compatibility issues.

Run:

```bash
./xsh.exe -i test_ssh.txt
```

**Expected**: Output should contain "hello world" if handshake succeeds.

### 4.2 SSH connection error

Use an unreachable port:

```
@ssh: host 10.255.255.1 command whoami
```

**Expected**: Should show connection timeout/refused error.

---

## 5. gRPC Execution

> **Note**: gRPC execution requires a running gRPC mock server (port 18081).

### 5.1 Health check

Create `test_grpc.txt`:

```
@grpc: HealthCheck
```

**Expected**: Should receive health response from the mock gRPC server.

---

## 6. LLM Mode

### 6.1 Mock LLM inference

```bash
./xsh.exe -m test-model -p "What is xsh?"
```

**Expected**: Returns mock LLM response for any prompt.

### 6.2 Auto-analyze task file with LLM

Create `test_migration.md`:

```markdown
# Migration Steps
The following steps should be executed in order:

1. Health check the production API
2. Wait 10 minutes
3. Ask user to confirm deployment
4. Deploy
5. Verify deployment
```

Run:

```bash
./xsh.exe -i test_migration.md -m test-model
```

**Expected**: The tool analyzes the file content and extracts structured tasks (GET health, wait, ask, POST deploy, check).

---

## 7. End-to-End Migration Flow

### 7.1 Full migration test

Create `test_full.txt`:

```
[GET] http://localhost:18080/health
@wait: 2min
@ask: 确认服务正常?
[POST] http://localhost:18080/post
@check: 验证部署结果
```

> Requires LLM model for @ask and @check. If no model, uses mock responses.

Run:

```bash
./xsh.exe -i test_full.txt
```

**Expected**:
- `[GET] health` returns 200 "healthy"
- `[WAIT] 2min` shows 2m0s
- `[ASK] 确认服务正常?` shows suggestion
- `[POST] /post` returns 201 "created"
- `[CHECK] 验证部署结果` shows PASS or FAIL

### 7.2 Migration with output file

```bash
./xsh.exe -i test_full.txt -o test_result.txt
cat test_result.txt
```

**Expected**: Result file contains execution output for all tasks.

---

## 8. CLI Flag Combinations

### 8.1 Stream mode

```bash
./xsh.exe -m test-model -p "Hello" -stream
```

**Expected**: Output is streamed line by line.

### 8.2 Input + Model + Output

```bash
./xsh.exe -i test_basic.txt -m any-model -o result.txt
cat result.txt
```

**Expected**: Tasks are executed and results written to `result.txt`.

---

## 9. Cleanup

After testing:

```bash
# Stop mock servers (Ctrl+C in each terminal)
# Remove temp files
rm -f test_basic.txt test_mixed.txt test_invalid.txt test_retry.txt
rm -f test_fail.txt test_slow.txt test_no_conn.txt
rm -f test_ssh.txt test_grpc.txt test_migration.md test_full.txt
rm -f test_result.txt result.txt
```

---

## 10. Known Issues

| Issue | Description | Workaround |
|-------|-------------|-----------|
| SSH kexLoop stall | Go 1.25 + `golang.org/x/crypto v0.50.0` SSH handshake hangs in kexLoop | Use Go 1.24 or skip SSH tests |
| gRPC WithBlock hang | `grpc.DialContext` + `WithBlock()` may not respect context timeout on Go 1.25 | Set explicit context timeout or skip |
| structpb/proto mismatch | `invokeGRPCMethod` uses `structpb.Value`, mock server expects typed proto messages | Rewrite executor to use typed client or rewrite mock to accept structpb |
| Network proxy required | Some external API calls (HuggingFace) may timeout without proxy | Set `HTTP_PROXY`/`HTTPS_PROXY` env vars |
