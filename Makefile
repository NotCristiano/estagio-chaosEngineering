SHELL := /bin/sh # Use sh for better compatibility across environments (for example if someone, me, is using fish)

DOCKERFILE := docker/Dockerfile # Path to the Dockerfile
IMAGE := chaos-ollama-poker:latest # Name and tag for the Docker image
IMAGE_TAR := /tmp/chaos-ollama-poker.tar # Temp path for the Docker build output
NAMESPACE := chaos-production # Kubernetes namespace for the application
MON_NS := monitoring # Kubernetes namespace for monitoring stack (Prometheus/Grafana)
MESH_NS := chaos-testing # Kubernetes namespace for Chaos Mesh
GRAFANA_URL := http://localhost:30300 # Local URL for Grafana (port forwarding assumed)
OLLAMA_IMAGE := chaos-ollama-model:latest # Name and tag for the Ollama Docker image

# Where we're telling Make that the names bellow are not files
.PHONY: helm-setup poke build-poker deploy-poker-k8s delete-poker-k8s logs build-docker restart-poker-k8s kill-ollama logs-ollama logs-ollama-init pause resume teardown

# Block that sets up helm related repos
helm-setup:
	@echo "Setting up Prometheus, Grafana and Chaos Mesh Helm repositories..."
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm repo add chaos-mesh https://charts.chaos-mesh.org
	helm repo update
	helm upgrade --install prometheus prometheus-community/kube-prometheus-stack -n $(MON_NS) --create-namespace -f k8s/prometheus-values.yaml
	helm upgrade --install chaos-mesh chaos-mesh/chaos-mesh -n $(MESH_NS) --create-namespace -f k8s/chaos-mesh-values.yaml
	@echo "Helm repositories ready."

build-ollama:
	@echo "Building Ollama image..."
	docker build --pull -f docker/ollama.Dockerfile -t ${OLLAMA_IMAGE} .

build-docker:
	@echo "Building Docker image: $(IMAGE)"
	docker build -f $(DOCKERFILE) -t $(IMAGE) .
	@echo "[OK] Docker image built: $(IMAGE)"

# Deploy poker and monitoring to Kubernetes
deploy-k8s: helm-setup build-docker build-ollama
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
	@echo "To view logs: make logs"

logs:
	@POD=$$(kubectl -n $(NAMESPACE) get pod -l app=chaos-poker -o jsonpath='{.items[0].metadata.name}' 2>/dev/null); \
	if [ -z "$$POD" ]; then \
		echo "No poker pod found in $(NAMESPACE)"; \
	else \
		echo "Tailing logs from $$POD..."; \
		kubectl -n $(NAMESPACE) logs -f $$POD; \
	fi

logs-ollama:
	kubectl logs -f -n $(NAMESPACE) -l app=ollama -c ollama

kill-ollama:
	@echo "Killing Ollama pod in $(NAMESPACE)..."
	kubectl -n $(NAMESPACE) delete pod -l app=ollama 2>/dev/null || echo "No Ollama pod found"
	@echo "[OK] Ollama pod killed (Kubernetes will restart it)"


teardown:
	@echo "Tearing down the environment..."
	docker compose -f docker/docker-compose.yml down 2>/dev/null || true

	@echo "Uninstalling Helm releases..."
	heml uninstall prometheus -n $(MON_NS) 2>/dev/null || true
	helm uninstall chaos-mesh -n $(MESH_NS) 2>/dev/null || true

	@echo "Removing Chaos Mesh CRDs..."
	kubectl delete crd --selector=app.kubernetes.io/instance=chaos-mesh 2>/dev/null || true

	@echo "Deleting Kubernetes namespaces..."
	kubectl delete namespace $(NAMESPACE) --ignore-not-found=true --timeout=60s || \
		kubectl get namespace $(NAMESPACE) -o json | \
		jq '.spec.finalizers = []' | \
		kubectl replace --raw "/api/v1/namespaces/$(NAMESPACE)/finalize" -f - 2>/dev/null || true

	kubectl delete namespace $(MON_NS) --ignore-not-found=true --timeout=60s || \
		kubectl get namespace $(MON_NS) -o json | \
		jq '.spec.finalizers = []' | \
		kubectl replace --raw "/api/v1/namespaces/$(MON_NS)/finalize" -f - 2>/dev/null || true

	kubectl delete namespace $(MESH_NS) --ignore-not-found=true --timeout=60s || \
		kubectl get namespace $(MESH_NS) -o json | \
		jq '.spec.finalizers = []' | \
		kubectl replace --raw "/api/v1/namespaces/$(MESH_NS)/finalize" -f - 2>/dev/null || true

	# Now cleanup tries for 60 seconds, if it fails it will force delete the namespaces by removing finalizers
	@echo "[OK] Everything is cleaned up."


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