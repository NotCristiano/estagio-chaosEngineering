SHELL := /bin/sh # Use sh for better compatibility across environments (for example if someone, me, is using fish)

DOCKERFILE := docker/Dockerfile # Path to the Dockerfile
IMAGE := chaos-ollama-poker:latest # Name and tag for the Docker image
IMAGE_TAR := /tmp/chaos-ollama-poker.tar # Temp path for the Docker build output
NAMESPACE := chaos-production # Kubernetes namespace for the application
MON_NS := monitoring # Kubernetes namespace for monitoring stack (Prometheus/Grafana)

# Where we're telling Make that the names bellow are not files
.PHONY: run-no-monitor run-with-monitor image-build image-import-k3s k8s-apply monitoring-up

# Only docker
run-no-monitor:
	docker compose -f docker/docker-compose.yml up

# Prometheus + Grafana
run-with-monitor: image-import-k3s k8s-apply monitoring-up

image-build:
	docker build -t $(IMAGE) -f $(DOCKERFILE) .

image-import-k3s: image-build
	docker save $(IMAGE) -o $(IMAGE_TAR)
	sudo k3s ctr images import $(IMAGE_TAR)

k8s-apply:
	kubectl apply -f k8s/ollama.yaml
	kubectl apply -f k8s/poker-deployment.yaml
	kubectl rollout restart deployment/chaos-poker -n $(NAMESPACE)

monitoring-up:
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm repo update
	helm upgrade --install prometheus prometheus-community/kube-prometheus-stack -n $(MON_NS) --create-namespace -f k8s/prometheus-values.yaml


