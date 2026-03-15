APP_NAME     := llmgopher
IMAGE        := $(APP_NAME):local
GATEWAY_PORT := 8080

# Kubernetes (production-like) settings
CLUSTER_NAME := $(APP_NAME)
KIND_CONFIG  := k8s/local/kind-config.yaml
K8S_DIR      := k8s/local

.PHONY: help dev dev-down dev-logs dev-ui-logs dev-ps dev-restart \
        build run test test-v \
        docker-build \
        k8s-up k8s-down k8s-deploy k8s-undeploy k8s-logs k8s-status \
        clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# ---------------------------------------------------------------------------
# Local development (docker-compose)
# ---------------------------------------------------------------------------

dev: ## Start all services via docker-compose (build + up)
	docker compose up --build -d postgres redis gateway ui
	@echo ""
	@echo "============================================"
	@echo "  Gateway is live at http://localhost:$(GATEWAY_PORT)"
	@echo "  Control Plane UI is live at http://localhost:3000"
	@echo "  Test key: sk-test-key-1"
	@echo ""
	@echo "  Example:"
	@echo "    curl http://localhost:$(GATEWAY_PORT)/health"
	@echo "============================================"

dev-down: ## Stop and remove all containers
	docker compose down

dev-logs: ## Tail logs from all services
	docker compose logs -f

dev-ui-logs: ## Tail logs from the UI service
	docker compose logs -f ui

dev-ps: ## Show running service status
	docker compose ps

dev-restart: ## Rebuild and restart gateway + UI
	docker compose up --build -d gateway ui

# ---------------------------------------------------------------------------
# Run locally (no containers — requires postgres + redis running)
# ---------------------------------------------------------------------------

build: ## Build the gateway binary
	go build -trimpath -o bin/gateway ./cmd/gateway

run: build ## Build and run the gateway locally (reads config.yaml)
	./bin/gateway

# ---------------------------------------------------------------------------
# Docker
# ---------------------------------------------------------------------------

docker-build: ## Build the gateway Docker image
	docker build -t $(IMAGE) .

# ---------------------------------------------------------------------------
# Kubernetes (kind — for production-like testing)
# ---------------------------------------------------------------------------

k8s-up: ## Create the kind cluster
	@if kind get clusters 2>/dev/null | grep -q "^$(CLUSTER_NAME)$$"; then \
		echo "Cluster '$(CLUSTER_NAME)' already exists"; \
	else \
		kind create cluster --name $(CLUSTER_NAME) --config $(KIND_CONFIG); \
	fi

k8s-down: ## Delete the kind cluster
	kind delete cluster --name $(CLUSTER_NAME)

k8s-deploy: docker-build ## Build image, load into kind, and deploy manifests
	kind load docker-image $(IMAGE) --name $(CLUSTER_NAME)
	kubectl apply -f $(K8S_DIR)/secrets.yaml
	kubectl apply -f $(K8S_DIR)/postgres.yaml
	kubectl apply -f $(K8S_DIR)/redis.yaml
	@echo "Waiting for Postgres to be ready..."
	kubectl rollout status deployment/postgres --timeout=90s
	@echo "Waiting for Redis to be ready..."
	kubectl rollout status deployment/redis --timeout=60s
	kubectl apply -f $(K8S_DIR)/gateway.yaml
	@echo "Waiting for Gateway to be ready..."
	kubectl rollout status deployment/gateway --timeout=90s

k8s-undeploy: ## Delete all K8s resources
	-kubectl delete -f $(K8S_DIR)/gateway.yaml
	-kubectl delete -f $(K8S_DIR)/redis.yaml
	-kubectl delete -f $(K8S_DIR)/postgres.yaml
	-kubectl delete -f $(K8S_DIR)/secrets.yaml

k8s-logs: ## Tail gateway logs in the kind cluster
	kubectl logs -f deployment/gateway

k8s-status: ## Show pod status in the kind cluster
	kubectl get pods -o wide

# ---------------------------------------------------------------------------
# Tests
# ---------------------------------------------------------------------------

test: ## Run the Go unit test suite
	go test ./... -count=1

test-v: ## Run tests with verbose output
	go test ./... -count=1 -v

# ---------------------------------------------------------------------------
# Cleanup
# ---------------------------------------------------------------------------

clean: dev-down ## Stop compose, delete kind cluster, remove build artifacts
	-kind delete cluster --name $(CLUSTER_NAME) 2>/dev/null
	-docker rmi $(IMAGE) 2>/dev/null
	-rm -rf bin/
