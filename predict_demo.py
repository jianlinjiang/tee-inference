#!/usr/bin/python3
import argparse
import json
import os
import sys
import time
import requests
import logging
import subprocess

log_file_path = '/home/admin/workspace/job/logs/user.log'  # predict日志路径

OLLAMA_API_URL = "http://localhost:80/v1/chat/completions"
MODEL_NAME = "Qwen3-1.7B-Q2_K.gguf"

def parse_arguments():
    """解析命令行参数"""
    parser = argparse.ArgumentParser(description='数据预测生成程序')
    parser.add_argument('--dataset', type=str, required=True,
                      help='输入数据集文件路径（JSONL格式）')
    parser.add_argument('--predictions', type=str, required=True,
                      help='预测结果输出路径（JSONL格式）')
    return parser.parse_args()

def start_ollama():
    # 启动 Llama Cpp 服务（后台运行）
    model_path = "/app/models/" + MODEL_NAME
    process = subprocess.Popen(
        ["/app/llama-server", "-c", "4096", "-b", "4096", "-t", "6",
		"-m", model_path, "--host", "0.0.0.0", "--port", "80"],
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT
        #text=True
    )

    # 等待服务就绪（最多 15 秒）
    start_time = time.time()
    while time.time() - start_time < 60:
        try:
            response = requests.get("http://localhost/health")
            if response.status_code == 200:
                logging.info("Llama cpp 服务已启动")
                return process
        except:
            time.sleep(1)

    raise Exception("Ollama 服务启动超时")

def stop_ollama(process):
    process.terminate()
    process.wait()
    logging.info("Llama cpp 服务已停止")

# 发送请求到 Ollama 并获取响应
def query_ollama(question):
    payload = {
        "timings_per_token": True,
        "model": MODEL_NAME,
        "messages": [
            {
                "role": "user",
                "content": question
            }
        ],
        "stream": False,  # 设置为 False 以一次性获取完整响应
    }
    logging.info(f"Payload {payload}")
    response = requests.post(OLLAMA_API_URL, json=payload)

    if response.status_code == 200:
        logging.info(f"Response {response.json()}")
        response_data = response.json()
        generated_text = response_data["choices"][0]["message"]["content"]
        eval_count = response_data["usage"]["total_tokens"]
        timings = response_data["timings"]
        return generated_text, eval_count, timings["prompt_ms"] + timings["predicted_ms"]
    else:
        logging.info(f"Error querying Ollama: {response.status_code}, {response.text}")
        return None, None, None

def main():
    os.makedirs(os.path.dirname(log_file_path), exist_ok=True)  # 确保日志目录存在

    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(levelname)s - %(message)s',
        handlers=[
            logging.FileHandler(log_file_path),
            logging.StreamHandler()  # 同时输出到控制台
        ]
    )

    args = parse_arguments()
    print(args)

    # 启动ollama服务
    ollama_process = start_ollama()

    prompt_count = 0
    results = []
    total_count = 0.0
    total_duration = 0.0
    with open(args.dataset, 'r', encoding='utf-8') as file:
        for line in file:
            prompt_count += 1

            data = json.loads(line.strip())
            prompt = data.get("input_text")
            result, count, duration = query_ollama(prompt)
            logging.info(f"ollama execution finished, result {result}, count {count}, duration {duration}")
            if not result == None:
                results.append(result)
                total_count += count
                total_duration += duration
            else:
                stop_ollama(ollama_process)
                sys.exit(1)

    # 生成预测结果
    prediction = {
        "result": results,
        "tokens per second": total_count * (10**3) / total_duration
    }

    predictions_dir = os.path.dirname(args.predictions)
    if predictions_dir and not os.path.exists(predictions_dir):
        try:
            os.makedirs(predictions_dir)
            # print(f"Created directory for predictions: {predictions_dir}")
            logging.info(f"Created directory for predictions: {predictions_dir}")
        except Exception as e:
            # print(f"Failed to create directory {predictions_dir}: {e}", file=sys.stderr)
            logging.info(f"Failed to create directory {predictions_dir}: {e}", file=sys.stderr)
            sys.exit(1)

    # 写入结果文件
    with open(args.predictions, 'w') as outfile:
        outfile.write(json.dumps(prediction) + '\n')

    # print(f"Result saved to {args.predictions}")
    logging.info(f"Result saved to {args.predictions}")

    # 停止ollama服务
    stop_ollama(ollama_process)

if __name__ == "__main__":
    main()