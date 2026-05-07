#!/usr/bin/env python3
"""
ONNX Runtime GenAI 测试推理脚本
使用 DeepSeek-R1-Distill-ONNX 模型进行推理
"""

import os

model_path = "models/deepseek-r1-distill-qwen-1.5B/cpu_and_mobile/cpu-int4-rtn-block-32-acc-level-4"

try:
    import onnxruntime_genai as og

    print(f"Loading model from: {model_path}")

    # 使用 Model 类
    model = og.Model(model_path)
    print("Model loaded successfully!")

    # 获取 tokenizer
    tokenizer = og.Tokenizer(model)

    # 准备输入
    prompt = "Hello, how are you?"

    # Tokenize
    input_ids = tokenizer.encode(prompt)
    print(f"Input tokens: {input_ids}")

    # 创建生成参数
    params = og.GeneratorParams(model)
    params.set_search_options(max_length=100)
    params.input_ids = input_ids

    # 创建生成器
    generator = og.Generator(model, params)

    # 生成
    print("\nGenerating...")
    generator.generate()

    # 获取输出
    output_tokens = generator.get_output(0)
    output_text = tokenizer.decode(output_tokens)

    print("\n" + "="*50)
    print("Input:", prompt)
    print("="*50)
    print("\nOutput:", output_text)
    print("="*50)

except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()