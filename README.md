# Chaos Engineering Poker

Este projeto tem dois modos de execucao pelo `Makefile`:

- `make run-no-monitor`: Docker local, sem Prometheus/Grafana.
- `make run-with-monitor`: k3s com app + Prometheus + Grafana.

## Primeiro setup

Faz este setup uma vez (ou quando mudares de maquina/ambiente).

### 1) Pre-requisitos

- Docker + Docker Compose
- `kubectl`
- `helm`
- `k3s` ativo (para o modo com monitoracao)

Verificacao rapida:

```fish
docker --version
docker compose version
kubectl version --client
helm version
kubectl get nodes
```

### 2) Setup inicial do modelo no Docker (modo local)

Inicia o Ollama e, numa segunda terminal, puxa o modelo:

```fish
cd /home/cristiano/GolandProjects/estagio-chaosEngineering
make run-no-monitor
```

```fish
docker exec -it ollama ollama pull llama3.2
```

### 3) Setup inicial do modelo no k3s (modo monitorado)

Sobe stack completa no cluster:

```fish
cd /home/cristiano/GolandProjects/estagio-chaosEngineering
make run-with-monitor
```

Quando o pod do Ollama estiver `Running`, puxa o modelo dentro do cluster:

```fish
kubectl exec -n chaos-production (kubectl get pod -n chaos-production -l app=ollama -o jsonpath='{.items[0].metadata.name}') -- ollama pull llama3.2
```

## Uso diario apos setup

### Sem monitoracao (mais simples)

Usa quando queres apenas testar o servico localmente:

```fish
cd /home/cristiano/GolandProjects/estagio-chaosEngineering
make run-no-monitor
```

### Com Prometheus + Grafana (k3s)

Usa quando queres observar metricas, disponibilidade e latencia no Grafana/Prometheus:

```fish
cd /home/cristiano/GolandProjects/estagio-chaosEngineering
make run-with-monitor
```

## Acesso a monitoracao

- Grafana e Prometheus sao instalados via Helm usando `k8s/prometheus-values.yaml`.
- Para descobrir o endpoint do Grafana no teu cluster:

```fish
kubectl get svc -n monitoring | grep grafana
```

- Para inspecionar targets no Prometheus (opcional):

```fish
kubectl -n monitoring port-forward svc/prometheus-kube-prometheus-prometheus 9090:9090
```

## Comandos uteis de diagnostico

```fish
kubectl get pods -n chaos-production
kubectl get pods -n monitoring
kubectl get servicemonitor -n chaos-production
kubectl logs -n chaos-production -l app=chaos-poker --tail=100
kubectl logs -n chaos-production -l app=ollama --tail=100
```
