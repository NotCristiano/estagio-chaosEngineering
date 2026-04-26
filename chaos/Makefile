.PHONY: setup kill-ollama run-latency report clean

setup:
	@echo "Setting up the environment and installing dependencies..."
	uv venv
	uv pip install chaostoolkit chaostoolkit-kubernetes chaostoolkit-prometheus chaostoolkit-reporting chaostoolkit-lib[jsonpath]
	@echo "[OK] Dependencies installed successfully!"

kill-ollama:
	@echo "Running kill-ollama experiment..."
	uv run chaos run experiments/kill-ollama.yaml

run-latency:
	@echo "Running latency-ollama experiment..."
	uv run chaos run experiments/latency-ollama.yaml

run-loss:
	@echo "Running network-loss experiment..."
	uv run chaos run experiments/loss-ollama.yaml

# WARNING: pandoc must be installed on the system for this to work
report:
	@echo "Generating HTML report from journal.json..."
	uv run chaos report --export-format=html journal.json report.html

clean:
	rm -rf .venv journal.json report.html