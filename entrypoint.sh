#!/bin/sh
# This entrypoint script checks if the whisper model file exists in the specified path.
#
MODEL_PATH=${WHISPER_MODEL_PATH:-/models/ggml-small.bin}
MODEL_DIR=$(dirname "$MODEL_PATH")
MODEL_FILE=$(basename "$MODEL_PATH")

mkdir -p "$MODEL_DIR"

if [ ! -f "$MODEL_PATH" ]; then
  echo "Downloading whisper model: $MODEL_FILE ..."
  curl -L "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/${MODEL_FILE}" \
    -o "$MODEL_PATH"
fi

exec /server
