SHELL := /bin/sh # Use sh for better compatibility across environments (for example if someone, me, is using fish)

DOCKERFILE := docker/Dockerfile # Path to the Dockerfile
IMAGE := chaos-ollama-poker:latest # Name and tag for the Docker image
IMAGE_TAR := /tmp/chaos-ollama-poker.tar # Temp path for the Docker build output
NAMESPACE := chaos-production # Kubernetes namespace for the application
MON_NS := monitoring # Kubernetes namespace for monitoring stack (Prometheus/Grafana)
GRAFANA_URL := http://localhost:30300 # Local URL for Grafana (port forwarding assumed)
OLLAMA_IMAGE := chaos-ollama-model:latest # Name and tag for the Ollama Docker image

#TODO: REWORK MAKEFILE TO MAKE IT PRETTIER AND LESS CLUSTERED

# Where we're telling Make that the names bellow are not files
.PHONY: run-no-monitor run-with-monitor monitoring-setup poke build-poker deploy-poker-k8s delete-poker-k8s logs-poker build-docker restart-poker-k8s kill-ollama logs-ollama logs-ollama-init pause resume

# run-with-monitor and run-no-monitor are now kinda deprecated since I remembered that I needed to use k8s, but still nice to have I guess

# Only docker
run-no-monitor:
	docker compose -f docker/docker-compose.yml up

# Docker with Prometheus + Grafana monitoring
run-with-monitor: monitoring-setup
	@echo "Grafana is available at: $(GRAFANA_URL)"
	@echo "Use admin / chaos123 to log in"
	docker compose -f docker/docker-compose.yml up

# Block that sets up Prometheus and Grafana using Helm
monitoring-setup:
	@echo "Setting up Prometheus + Grafana"
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm repo update
	helm upgrade --install prometheus prometheus-community/kube-prometheus-stack -n $(MON_NS) --create-namespace -f k8s/prometheus-values.yaml
	@echo "Monitoring stack ready."

# Helper that builds the poker binary
build-poker:
	@echo "Building poker binary..."
	CGO_ENABLED=0 go build -o poker ./cmd/poker

build-ollama:
	@echo "Building Ollama image..."
	docker build -f docker/ollama.Dockerfile -t ${OLLAMA_IMAGE} .

# Run the poker binary with pre-specified parameters
poke: build-poker
	@echo "Starting poker"
	./poker --target=http://localhost:11434 --model=llama3.2 --count=0 --interval=5s --timeout=60s --retries=3 --metric-addr=:9090

# Build Docker image for the poker app
build-docker:
	@echo "Building Docker image: $(IMAGE)"
	docker build -f $(DOCKERFILE) -t $(IMAGE) .
	@echo "[OK] Docker image built: $(IMAGE)"

# Deploy poker and monitoring to Kubernetes
deploy-k8s: monitoring-setup build-docker build-ollama
	@echo "Importing image into k3s..."
	docker save $(IMAGE) | sudo k3s ctr images import -
	docker save $(OLLAMA_IMAGE) | sudo k3s ctr images import -
	@echo "Deploying to Kubernetes namespace $(NAMESPACE)..."
	kubectl create namespace $(NAMESPACE) 2>/dev/null || true
	kubectl apply -f k8s/ollama.yaml
	kubectl apply -f k8s/poker-deployment.yaml
	kubectl apply -f k8s/dashboards/
	@echo "[OK] Deployed to $(NAMESPACE)"
	@echo "To check poker pod status: kubectl -n $(NAMESPACE) get pods -l app=chaos-poker"
	@echo "To view logs: make logs-poker"

# Show poker logs in Kubernetes
logs-poker:
	@POD=$$(kubectl -n $(NAMESPACE) get pod -l app=chaos-poker -o jsonpath='{.items[0].metadata.name}' 2>/dev/null); \
	if [ -z "$$POD" ]; then \
		echo "No poker pod found in $(NAMESPACE)"; \
	else \
		echo "Tailing logs from $$POD..."; \
		kubectl -n $(NAMESPACE) logs -f $$POD; \
	fi

logs-ollama:
	kubectl logs -f -n $(NAMESPACE) -l app=ollama -c ollama

logs-ollama-init:
	kubectl logs -f -n $(NAMESPACE) -l app=ollama -c model-puller

# Kill Ollama pod (chaos testing - forces restart)
kill-ollama:
	@echo "Killing Ollama pod in $(NAMESPACE)..."
	kubectl -n $(NAMESPACE) delete pod -l app=ollama 2>/dev/null || echo "No Ollama pod found"
	@echo "[OK] Ollama pod killed (Kubernetes will restart it)"


# Completely tears down the project and stops all background processes
teardown:
	@echo "Tearing down the environment..."
	# Stop any Docker Compose containers
	docker compose -f docker/docker-compose.yml down 2>/dev/null || true
	# Gracefully uninstall Helm charts
	helm uninstall prometheus -n $(MON_NS) 2>/dev/null || true
	# Delete the Kubernetes namespaces
	kubectl delete namespace $(NAMESPACE) $(MON_NS) --ignore-not-found=true
	@echo "Everything is cleaned up."

pause:
	@echo "Pausing all workloads in $(NAMESPACE)..."
	kubectl scale deployment/ollama -n $(NAMESPACE) --replicas=0
	kubectl scale deployment/chaos-poker -n $(NAMESPACE) --replicas=0
	kubectl scale -n $(MON_NS) \
		deployment/prometheus-grafana \
		deployment/prometheus-kube-prometheus-operator \
		deployment/prometheus-kube-state-metrics \
		--replicas=0
	kubectl scale statefulset/prometheus-prometheus-kube-prometheus-prometheus -n $(MON_NS) --replicas=0
	@echo "[OK] All workloads paused. Data and config preserved."

resume:
	@echo "Resuming all workloads..."
	kubectl scale deployment/ollama -n $(NAMESPACE) --replicas=1
	kubectl scale deployment/chaos-poker -n $(NAMESPACE) --replicas=1
	kubectl scale -n $(MON_NS) \
		deployment/prometheus-grafana \
		deployment/prometheus-kube-prometheus-operator \
		deployment/prometheus-kube-state-metrics \
		--replicas=1
	kubectl scale statefulset/prometheus-prometheus-kube-prometheus-prometheus -n $(MON_NS) --replicas=1
	@echo "[OK] Workloads resumed."cho "[OK] Workloads resumed."