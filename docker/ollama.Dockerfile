# FROM ollama/ollama:latest #CPU only
FROM ollama/ollama:rocm

RUN ollama serve & until ollama list >/dev/null 2>&1; do sleep 1; done && ollama pull qwen3.5:2b