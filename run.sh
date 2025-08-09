#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PARENT_DIR="$(dirname "$SCRIPT_DIR")"

# 根据运行环境选择文件路径
if [ "$ALIPAY_APP_ENV" = "prod" ]; then
    PREDICTIONS_RESULT_FILE="/home/admin/workspace/job/output/predictions/predictions.jsonl"
    DATASET_FILE="/home/admin/workspace/job/input/test.jsonl"
else
    PREDICTIONS_RESULT_FILE="${PARENT_DIR}/data/predictions.jsonl"
    DATASET_FILE="${PARENT_DIR}/data/test.jsonl"
fi

# 执行预测代码 ## 可修改为任意实现
SCRIPT_DIR=$(dirname "$0")
chmod 777 "${SCRIPT_DIR}/predict_demo.py"
"${SCRIPT_DIR}/predict_demo.py" \
    --dataset "$DATASET_FILE" \
    --predictions "$PREDICTIONS_RESULT_FILE"