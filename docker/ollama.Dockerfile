FROM ollama/ollama:latest

RUN ollama serve & until ollama list >/dev/null 2>&1; do sleep 1; done && ollama pull qwen3.5:2b