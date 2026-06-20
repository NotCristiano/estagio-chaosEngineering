# estagio-chaosEngineering

Chaos engineering project for an Ollama-based question-answering service.

The Go application sends probes to Ollama, exposes Prometheus metrics, and can run periodic quality checks with the ARC dataset. The repository also includes Kubernetes manifests, monitoring setup, and Chaos Toolkit experiments.

## Structure

- `cmd/poker`: Go application entrypoint
- `internal`: Ollama client, metrics, dataset loading, reporting
- `k8s`: Kubernetes manifests and Grafana dashboards
- `docker`: container definitions
- `chaos`: Chaos Toolkit experiments and reporting commands

## Prerequisites

- `go`
- `make`
- `docker`
- `kubectl`
- `helm`
- `k3s`
- `uv` for the commands under `chaos/`

## Main workflow

Deploy the full environment:

```bash
make deploy-k8s
```

Tail application logs:

```bash
make logs
```

Pause or resume the environment:

```bash
make pause
make resume
```

Remove the environment:

```bash
make teardown
```

## Root Makefile

Relevant commands from the main `Makefile`:

- `make deploy-k8s`: builds the images, installs monitoring and Chaos Mesh, and deploys the app to Kubernetes
- `make logs`: tails the `chaos-poker` application logs
- `make logs-ollama`: tails the Ollama pod logs
- `make kill-ollama`: deletes the Ollama pod so Kubernetes recreates it
- `make pause`: scales the workloads down to zero
- `make resume`: scales the workloads back up
- `make teardown`: removes the deployed environment

## chaos/Makefile

Available commands in `chaos/Makefile`:

- `make -C chaos setup`
- `make -C chaos run-baseline`
- `make -C chaos report-baseline`
- `make -C chaos run-stress-cpu`
- `make -C chaos run-latency`
- `make -C chaos run-loss`
- `make -C chaos run-dns`
- `make -C chaos run-oom`
- `make -C chaos run-disk`
- `make -C chaos run-scale`
- `make -C chaos run-concurrency`
- `make -C chaos run-bad-request`
- `make -C chaos kill-ollama`
- `make -C chaos report-comparison`
- `make -C chaos clean`

## Notes

- The Kubernetes deployment uses the `chaos-production` namespace for the application.
- Monitoring is installed in `monitoring`.
- Chaos Mesh is installed in `chaos-testing`.
