.PHONY: setup run-baseline run-stress-cpu run-latency run-loss run-dns run-oom run-disk run-scale run-concurrency run-bad-request kill-ollama report-baseline report-comparison clean

GENERATED_JOURNALS := $(wildcard *__journal.json)
BASELINE_JOURNAL := $(wildcard baseline_journal.json)

setup:
	@echo "Setting up the environment and installing dependencies..."
	uv venv
	uv pip install chaostoolkit chaostoolkit-kubernetes chaostoolkit-prometheus chaostoolkit-reporting chaostoolkit-lib[jsonpath]
	@echo "[OK] Dependencies installed successfully!"

# BASELINE RUNNERS
run-baseline:
	@echo "Running baseline-metrics..."
	uv run chaos run experiments/baseline-metrics.yaml --journal-path=baseline_journal.json

report-baseline:
	@echo "Generating HTML baseline report..."
	uv run chaos report --export-format=html baseline_journal.json baseline_report.html

# EXPERIMENT RUNNERS
run-stress-cpu:
	@echo "Running cpu-hog experiment..."
	uv run chaos run experiments/cpu-stress-ollama.yaml --journal-path=cpu_stress__journal.json

run-latency:
	@echo "Running latency-ollama experiment..."
	uv run chaos run experiments/latency-ollama.yaml --journal-path=latency__journal.json

run-loss:
	@echo "Running network-loss experiment..."
	uv run chaos run experiments/loss-ollama.yaml --journal-path=network_loss__journal.json

run-dns:
	@echo "Running dns-failure experiment..."
	uv run chaos run experiments/dns-poker.yaml --journal-path=dns__journal.json

run-oom:
	@echo "Running memory-leak experiment..."
	uv run chaos run experiments/oom-ollama.yaml --journal-path=oom__journal.json

run-disk:
	@echo "Running disk-ollama experiment..."
	uv run chaos run experiments/disk-ollama.yaml --journal-path=disk__journal.json

run-scale:
	@echo "Running scale-poker experiment..."
	uv run chaos run experiments/scale-poker.yaml --journal-path=scale__journal.json

run-concurrency:
	@echo "Running concurrency-poker..."
	uv run chaos run experiments/concurrency-poker.yaml --journal-path=concurrency__journal.json

run-bad-request:
	@echo "Running bad-request-poker..."
	uv run chaos run experiments/bad-request-poker.yaml --journal-path=bad_request__journal.json

kill-ollama:
	@echo "Running kill-ollama experiment..."
	uv run chaos run experiments/kill-ollama.yaml --journal-path=kill_ollama__journal.json

# REPORTING & AUTOMATED COMPARISON
# Merges the baseline (if present) with any experiment journals currently existing in the workspace
report-comparison:
	@if [ -z "$(GENERATED_JOURNALS)" ]; then \
		echo "Error: No experiment journals found to report on. Run a test first!"; \
		exit 1; \
	fi
	@echo "Found baseline journal: $(BASELINE_JOURNAL)"
	@echo "Found active experiment journals: $(GENERATED_JOURNALS)"
	@echo "Generating unified comparison report reflecting differences..."
	uv run chaos report --export-format=html $(BASELINE_JOURNAL) $(GENERATED_JOURNALS) comparison_report.html
	@echo "[OK] Report compiled successfully as comparison_report.html!"

clean:
	rm -rf .venv baseline_journal.json *__journal.json baseline_report.html comparison_report.html