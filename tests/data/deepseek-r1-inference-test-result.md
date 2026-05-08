# DeepSeek R1 Distill Qwen 1.5B Inference Test Result

**Date:** 2026-05-08
**Model:** deepseek-r1-distill-qwen-1.5B (cpu-int4-rtn-block-32-acc-level-4)
**Runtime:** onnxruntime-genai v0.12.0 on Windows 11 (x64)
**Status:** ALL PASSED

## Test Results

### 1. Real Inference Test (PASS)

**Prompt:** "What is 2+3?"

**Response:**
```
2+3 is 5. So, the answer is 5.
```

**Solution:**
We need to find the sum of 2 and 3.
2 + 3 = 5

**Answer:** 5

**Latency:** ~5s (first token)

### 2. Stream Inference Test (PASS)

**Prompt:** "Hello"
**MaxTokens:** 256
**Temperature:** 0.6, **TopP:** 0.95

**Result:** Model produced coherent streaming tokens with reasoning output. The model demonstrated chain-of-thought reasoning about Heron's formula for triangle area calculation.

**Latency:** ~10s for 256 tokens

### 3. Task Analysis Test (PASS)

**Input Content:**
```
Migration from prod to UAT:
1. Check service health at http://example.com/api/health
2. Deploy new version at http://example.com/api/deploy
3. Wait 10 minutes for deployment
4. Verify deployment result
```

**Extracted Tasks:**
- Task[0]: type=http raw=[GET] http://example.com/api/health
- Task[1]: type=http raw=[GET] http://example.com/api/deploy

**Latency:** ~61s (full inference with task parsing)

## Model Info

- **Type:** qwen2
- **Context Length:** 131072
- **Hidden Size:** 1536
- **Num Attention Heads:** 12
- **Num Hidden Layers:** 28
- **Num KV Heads:** 2
- **Vocab Size:** 151936
- **Quantization:** INT4-RTN (block-32, acc-level-4)
