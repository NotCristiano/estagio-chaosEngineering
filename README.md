# Chaos Engineering Poker
We will use primarly the `Makefile` to run this project:

- `make run-no-monitor`: start Ollama locally with Docker Compose.
- `make run-with-monitor`: installs Prometheus + Grafana in k3s, then starts Ollama.
- `make poke`: runs the poker app.
- `make deploy-k8s`: deploys Ollama, Poker and the monitoring stack to k8s.
- `make kill-ollama`: kills the local Ollama in k8s so we can observe changes in grafana.

## Running the poker app

Run this to setup the repository on your environment:

### 1) Prerequisites

- Docker + Docker Compose
- `kubectl`
- `helm`
- `k3s` (or another Kubernetes cluster)

Quick check:

```fish
docker --version
docker compose version
kubectl version --client
helm version
kubectl get nodes
```

### 2) Local Ollama only

Start Ollama in the foreground:

```fish
make run-no-monitor
```

### 3) Ollama + monitoring

Install the monitoring stack in k3s and start Ollama in the foreground:

```fish
make run-with-monitor
```

This does two things:

1. Installs Prometheus + Grafana using `k8s/prometheus-values.yaml`.
2. Starts Ollama locally with Docker Compose.

If you need the model loaded into the local Ollama container, run the follow:

```fish
docker exec -it ollama ollama pull llama3.2
```

### 4) Running the poker


The poker app is separate from Ollama. It connects to the local Ollama instance at `http://localhost:11434`.

Start it with:

```fish
make poke
```

`make poke` builds the binary if needed and then runs it forever with:

- target: `http://localhost:11434`
- model: `llama3.2`
- interval: `5s`
- count: `0` (infinite)

Stop it with `Ctrl+C`.

### 5) Deploying everything to k8s
You can deploy Ollama, the poker app and the monitoring stack to k3s with:

```fish
make deploy-k8s
```

This will install the monitoring stack using `k8s/prometheus-values.yaml` and then deploy Ollama and the poker app.

### 6) Killing Ollama in k8s
You can kill Ollama in k8s with:
```fish
make kill-ollama
```
This will temporarily remove the Ollama deployment from k8s, that way we can observe changes in grafana. Like the success and failure rate.

### 7) Port forwarding (JUST IN CASE)
If for some reason it's not working for you, try to setup port forwarding, here's an example:

```fish
kubectl -n monitoring port-forward svc/prometheus-grafana 3000:80
kubectl -n monitoring port-forward svc/prometheus-kube-prometheus-prometheus 9090:9090
```
