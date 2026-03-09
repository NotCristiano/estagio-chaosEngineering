# How to use

## First Step

build the poker binary
```fish
cd estagio-chaosEngineering
go build -o poker ./cmd/poker
```

---

## Docker Test

Start ollama
```fish
cd docker
docker compose up -d
```

Pull the model (only need to do this once)
```fish
docker exec -it ollama ollama pull llama3.2
```

Run the poker
```fish
./poker --target http://localhost:11434 --model llama3.2 --count 10
```

Manual Chaos, on a second terminal kill ollama while the poker is running
```fish
docker stop ollama
docker start ollama
```

---

## Kubernetes Test

Make sure k3s (or the one of your choice) is running
```fish
kubectl get nodes
```

Put ollama into the cluster
```fish
kubectl apply -f k8s/ollama.yaml
kubectl get pods -n chaos-production -w
```

When it's running pull the model
```fish
kubectl exec -n chaos-production (kubectl get pod -n chaos-production -l app=ollama -o jsonpath='{.items[0].metadata.name}') -- ollama pull llama3.2
```

Run the poker against the cluster (different port than before)
```fish
./poker --target http://localhost:30434 --model llama3.2 --count 20
```

Semi-Automatic Chaos, kill the pod while poker is running, k8s will restart it
```fish
kubectl delete pod -n chaos-lab -l app=ollama
```

---

## Cheat Sheet

Check cluster is healthy
```fish
kubectl get pods -n kube-system
kubectl get pods -n chaos-lab
```

Check ollama logs
```fish
kubectl logs -n chaos-lab -l app=ollama
```

Start fresh
```fish
kubectl delete namespace chaos-lab
kubectl apply -f k8s/ollama.yaml
```

Poker flags
```
--count 0        infinite mode (ctrl+c to stop)
--interval 1s    how often to probe
--timeout 60s    per request timeout (bump this for cpu inference)
--retries 3      retries before marking as failed
```
