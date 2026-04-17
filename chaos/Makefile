.PHONY: setup run report clean

setup:
	@echo "Setting up the environment and installing dependencies..."
	uv venv
	uv pip install chaostoolkit chaostoolkit-kubernetes chaostoolkit-prometheus chaostoolkit-reporting chaostoolkit-lib[jsonpath]
	@echo "[OK] Dependencies installed successfully!"

# For now just the ollama kill experiment, but we will add more in the future
run:
	@echo "Running kill-ollama experiment..."
	uv run chaos run experiments/kill-ollama.yaml

# WARNING: pandoc must be installed on the system for this to work
report:
	@echo "Generating HTML report from journal.json..."
	uv run chaos report --export-format=html journal.json report.html

clean:
	rm -rf .venv journal.json report.html