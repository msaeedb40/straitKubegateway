# straitKubegateway Makefile
# ============================================================================

SHELL := /bin/bash
.DEFAULT_GOAL := help

# Project metadata
MODULE := github.com/straitkubegateway/straitkubegateway
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Go
GO ?= go
GOFLAGS ?= -trimpath
LDFLAGS := -s -w \
	-X $(MODULE)/internal/version.Version=$(VERSION) \
	-X $(MODULE)/internal/version.GitCommit=$(GIT_COMMIT) \
	-X $(MODULE)/internal/version.BuildDate=$(BUILD_DATE)

# Architectures
ARCH ?= $(shell go env GOARCH)
OS ?= linux

# Binaries
BIN_DIR := bin
STRAITD := $(BIN_DIR)/straitd
SG_CONTROLLER := $(BIN_DIR)/sg-controller
SG_CLI := $(BIN_DIR)/sg-cli

# BPF
CLANG ?= clang-22
LLC ?= llc-22
BPF_DIR := bpf
BPF_OUT := ebpf/generated
BPFTOOL ?= bpftool

# Docker
REGISTRY ?= ghcr.io/straitkubegateway
IMAGE_TAG ?= $(VERSION)

# Helm
HELM ?= helm
HELM_DIR := straitKubegateway-helm-repo

# ============================================================================
# Build
# ============================================================================

.PHONY: build
build: build-straitd build-sg-controller build-sg-cli ## Build all Go binaries

.PHONY: build-straitd
build-straitd: ## Build straitd node agent
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) $(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(STRAITD) ./cmd/straitd/

.PHONY: build-sg-controller
build-sg-controller: ## Build sg-controller
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) $(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(SG_CONTROLLER) ./cmd/sg-controller/

.PHONY: build-sg-cli
build-sg-cli: ## Build sg-cli
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) $(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o $(SG_CLI) ./cmd/sg-cli/

.PHONY: build-all-arch
build-all-arch: ## Build for amd64 and arm64
	ARCH=amd64 $(MAKE) build
	ARCH=arm64 $(MAKE) build

# ============================================================================
# BPF
# ============================================================================

.PHONY: bpf
bpf: ## Compile eBPF programs
	@mkdir -p $(BPF_OUT)
	@echo "Compiling eBPF programs..."
	@for src in $(BPF_DIR)/src/*.c; do \
		base=$$(basename $$src .c); \
		$(CLANG) -O2 -g -target bpf \
			-D__TARGET_ARCH_$(ARCH) \
			-I$(BPF_DIR)/headers \
			-c $$src \
			-o $(BPF_OUT)/$$base.o; \
		echo "  ✓ $$base.o"; \
	done

.PHONY: bpf-generate
bpf-generate: bpf ## Generate Go bindings from BPF objects
	$(GO) generate ./ebpf/...

# ============================================================================
# Generate
# ============================================================================

.PHONY: generate
generate: ## Run all code generation
	$(GO) generate ./...
	@echo "Running CRD generation..."
	hack/generate/generate-crds.sh
	hack/generate/generate-deepcopy.sh

.PHONY: manifests
manifests: ## Generate CRD manifests
	hack/generate/generate-crds.sh

# ============================================================================
# Test
# ============================================================================

.PHONY: test
test: ## Run unit tests
	$(GO) test -race -count=1 ./...

.PHONY: test-unit
test-unit: ## Run unit tests with coverage
	$(GO) test -race -coverprofile=coverage.out -count=1 ./...
	$(GO) tool cover -func=coverage.out

.PHONY: test-integration
test-integration: ## Run integration tests
	$(GO) test -race -tags=integration -count=1 ./test/integration/...

.PHONY: test-e2e
test-e2e: ## Run end-to-end tests
	hack/test/test-e2e.sh

.PHONY: test-dataplane
test-dataplane: ## Run dataplane tests (requires root + kernel >= 6.7)
	sudo $(GO) test -tags=dataplane -count=1 ./test/dataplane/...

.PHONY: test-bpf
test-bpf: ## Run BPF verifier tests
	sudo $(GO) test -tags=bpf -count=1 ./ebpf/...

# ============================================================================
# Lint & Verify
# ============================================================================

.PHONY: lint
lint: ## Run linters
	golangci-lint run ./...

.PHONY: fmt
fmt: ## Format Go code
	gofmt -s -w .

.PHONY: vet
vet: ## Run go vet
	$(GO) vet ./...

.PHONY: verify
verify: ## Run all verification checks
	hack/verify/verify-gofmt.sh
	hack/verify/verify-govet.sh
	hack/verify/verify-generated.sh
	hack/verify/verify-crds.sh
	hack/verify/verify-bpf.sh
	hack/verify/verify-manifests.sh

# ============================================================================
# Docker
# ============================================================================

.PHONY: docker
docker: docker-straitd docker-sg-controller docker-sg-cli docker-ui ## Build all Docker images

.PHONY: docker-straitd
docker-straitd: ## Build straitd Docker image
	docker buildx build --platform linux/amd64,linux/arm64 \
		-f build/Dockerfile.straitd \
		-t $(REGISTRY)/straitd:$(IMAGE_TAG) \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		.

.PHONY: docker-sg-controller
docker-sg-controller: ## Build sg-controller Docker image
	docker buildx build --platform linux/amd64,linux/arm64 \
		-f build/Dockerfile.sg-controller \
		-t $(REGISTRY)/sg-controller:$(IMAGE_TAG) \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		.

.PHONY: docker-sg-cli
docker-sg-cli: ## Build sg-cli Docker image
	docker buildx build --platform linux/amd64,linux/arm64 \
		-f build/Dockerfile.sg-cli \
		-t $(REGISTRY)/sg-cli:$(IMAGE_TAG) \
		--build-arg VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		.

.PHONY: docker-ui
docker-ui: ## Build UI Docker image
	docker buildx build --platform linux/amd64,linux/arm64 \
		-f build/Dockerfile.ui \
		-t $(REGISTRY)/ui:$(IMAGE_TAG) \
		.

.PHONY: docker-push
docker-push: ## Push all Docker images to registry
	docker push $(REGISTRY)/straitd:$(IMAGE_TAG)
	docker push $(REGISTRY)/sg-controller:$(IMAGE_TAG)
	docker push $(REGISTRY)/sg-cli:$(IMAGE_TAG)
	docker push $(REGISTRY)/ui:$(IMAGE_TAG)

# ============================================================================
# Kind & Local Deployment
# ============================================================================

KIND_CLUSTER ?= cluster0

.PHONY: kind-load
kind-load: ## Build and load all images (straitd, sg-controller, sg-cli, ui) into Kind cluster
	docker build -f build/Dockerfile.sg-controller -t $(REGISTRY)/sg-controller:$(IMAGE_TAG) .
	docker build -f build/Dockerfile.straitd -t $(REGISTRY)/straitd:$(IMAGE_TAG) .
	docker build -f build/Dockerfile.sg-cli -t $(REGISTRY)/sg-cli:$(IMAGE_TAG) .
	docker build -f build/Dockerfile.ui -t $(REGISTRY)/ui:$(IMAGE_TAG) .
	kind load docker-image $(REGISTRY)/sg-controller:$(IMAGE_TAG) --name $(KIND_CLUSTER)
	kind load docker-image $(REGISTRY)/straitd:$(IMAGE_TAG) --name $(KIND_CLUSTER)
	kind load docker-image $(REGISTRY)/sg-cli:$(IMAGE_TAG) --name $(KIND_CLUSTER)
	kind load docker-image $(REGISTRY)/ui:$(IMAGE_TAG) --name $(KIND_CLUSTER)

.PHONY: kind-deploy
kind-deploy: kind-load ## Build, load into Kind, and deploy via Helm
	$(HELM) upgrade --install straitkubegateway $(HELM_DIR) \
		--namespace kube-system \
		--create-namespace \
		--set kubeProxyReplacement=true \
		--set disableDefaultCNI=true \
		--set straitd.image.tag=$(IMAGE_TAG) \
		--set straitd.image.pullPolicy=IfNotPresent \
		--set sgController.image.tag=$(IMAGE_TAG) \
		--set sgController.image.pullPolicy=IfNotPresent \
		--set ui.image.tag=$(IMAGE_TAG) \
		--set ui.image.pullPolicy=IfNotPresent \
		--wait --timeout=15m

# ============================================================================
# Minikube Deployment
# ============================================================================

MINIKUBE_PROFILE ?= minikube

.PHONY: minikube-load
minikube-load: ## Build and load all images into Minikube cluster
	docker build -f build/Dockerfile.sg-controller -t $(REGISTRY)/sg-controller:$(IMAGE_TAG) .
	docker build -f build/Dockerfile.straitd -t $(REGISTRY)/straitd:$(IMAGE_TAG) .
	docker build -f build/Dockerfile.sg-cli -t $(REGISTRY)/sg-cli:$(IMAGE_TAG) .
	docker build -f build/Dockerfile.ui -t $(REGISTRY)/ui:$(IMAGE_TAG) .
	minikube image load $(REGISTRY)/sg-controller:$(IMAGE_TAG) -p $(MINIKUBE_PROFILE)
	minikube image load $(REGISTRY)/straitd:$(IMAGE_TAG) -p $(MINIKUBE_PROFILE)
	minikube image load $(REGISTRY)/sg-cli:$(IMAGE_TAG) -p $(MINIKUBE_PROFILE)
	minikube image load $(REGISTRY)/ui:$(IMAGE_TAG) -p $(MINIKUBE_PROFILE)

.PHONY: minikube-deploy
minikube-deploy: minikube-load ## Build, load into Minikube, and deploy via Helm
	$(HELM) upgrade --install straitkubegateway $(HELM_DIR) \
		--namespace kube-system \
		--create-namespace \
		--set kubeProxyReplacement=true \
		--set disableDefaultCNI=true \
		--set straitd.image.tag=$(IMAGE_TAG) \
		--set straitd.image.pullPolicy=IfNotPresent \
		--set sgController.image.tag=$(IMAGE_TAG) \
		--set sgController.image.pullPolicy=IfNotPresent \
		--set ui.image.tag=$(IMAGE_TAG) \
		--set ui.image.pullPolicy=IfNotPresent \
		--wait --timeout=15m

# ============================================================================
# K3s / k3d Deployment
# ============================================================================

K3D_CLUSTER ?= straitkubegateway

.PHONY: k3d-load
k3d-load: ## Build and load all images into k3d cluster
	docker build -f build/Dockerfile.sg-controller -t $(REGISTRY)/sg-controller:$(IMAGE_TAG) .
	docker build -f build/Dockerfile.straitd -t $(REGISTRY)/straitd:$(IMAGE_TAG) .
	docker build -f build/Dockerfile.sg-cli -t $(REGISTRY)/sg-cli:$(IMAGE_TAG) .
	docker build -f build/Dockerfile.ui -t $(REGISTRY)/ui:$(IMAGE_TAG) .
	k3d image import $(REGISTRY)/sg-controller:$(IMAGE_TAG) --cluster $(K3D_CLUSTER)
	k3d image import $(REGISTRY)/straitd:$(IMAGE_TAG) --cluster $(K3D_CLUSTER)
	k3d image import $(REGISTRY)/sg-cli:$(IMAGE_TAG) --cluster $(K3D_CLUSTER)
	k3d image import $(REGISTRY)/ui:$(IMAGE_TAG) --cluster $(K3D_CLUSTER)

.PHONY: k3d-deploy
k3d-deploy: k3d-load ## Build, load into k3d, and deploy via Helm
	$(HELM) upgrade --install straitkubegateway $(HELM_DIR) \
		--namespace kube-system \
		--create-namespace \
		--set kubeProxyReplacement=true \
		--set disableDefaultCNI=true \
		--set straitd.image.tag=$(IMAGE_TAG) \
		--set straitd.image.pullPolicy=IfNotPresent \
		--set sgController.image.tag=$(IMAGE_TAG) \
		--set sgController.image.pullPolicy=IfNotPresent \
		--set ui.image.tag=$(IMAGE_TAG) \
		--set ui.image.pullPolicy=IfNotPresent \
		--wait --timeout=15m

.PHONY: k3s-load
k3s-load: ## Build and import images directly into local k3s containerd
	docker build -f build/Dockerfile.sg-controller -t $(REGISTRY)/sg-controller:$(IMAGE_TAG) .
	docker build -f build/Dockerfile.straitd -t $(REGISTRY)/straitd:$(IMAGE_TAG) .
	docker build -f build/Dockerfile.sg-cli -t $(REGISTRY)/sg-cli:$(IMAGE_TAG) .
	docker build -f build/Dockerfile.ui -t $(REGISTRY)/ui:$(IMAGE_TAG) .
	docker save $(REGISTRY)/sg-controller:$(IMAGE_TAG) | sudo k3s ctr images import -
	docker save $(REGISTRY)/straitd:$(IMAGE_TAG) | sudo k3s ctr images import -
	docker save $(REGISTRY)/sg-cli:$(IMAGE_TAG) | sudo k3s ctr images import -
	docker save $(REGISTRY)/ui:$(IMAGE_TAG) | sudo k3s ctr images import -

.PHONY: k3s-deploy
k3s-deploy: k3s-load ## Build, import into local k3s, and deploy via Helm
	$(HELM) upgrade --install straitkubegateway $(HELM_DIR) \
		--namespace kube-system \
		--create-namespace \
		--set kubeProxyReplacement=true \
		--set disableDefaultCNI=true \
		--set straitd.image.tag=$(IMAGE_TAG) \
		--set straitd.image.pullPolicy=IfNotPresent \
		--set sgController.image.tag=$(IMAGE_TAG) \
		--set sgController.image.pullPolicy=IfNotPresent \
		--set ui.image.tag=$(IMAGE_TAG) \
		--set ui.image.pullPolicy=IfNotPresent \
		--wait --timeout=15m

# ============================================================================
# Helm
# ============================================================================

.PHONY: helm-lint
helm-lint: ## Lint Helm chart
	$(HELM) lint $(HELM_DIR)

.PHONY: helm-template
helm-template: ## Render Helm chart templates
	$(HELM) template straitkubegateway $(HELM_DIR)

.PHONY: helm-package
helm-package: ## Package Helm chart
	$(HELM) package $(HELM_DIR) -d dist/charts

.PHONY: helm-index
helm-index: ## Build and index Helm repository for GitHub Pages (https://msaeedb40.github.io/straitKubegateway)
	@chmod +x scripts/publish-helm-repo.sh
	./scripts/publish-helm-repo.sh dist/charts

.PHONY: helm-serve
helm-serve: helm-index ## Start local HTTP server to test charts repository on port 8080
	@echo "Starting local chart repository server at http://localhost:8080 ..."
	@echo "Test with: helm repo add local-test http://localhost:8080 && helm repo update"
	python3 -m http.server 8080 --directory dist/charts


# ============================================================================
# Proto
# ============================================================================

.PHONY: proto
proto: ## Generate protobuf code
	buf generate proto

# ============================================================================
# Clean
# ============================================================================

.PHONY: clean
clean: ## Remove build artifacts
	rm -rf $(BIN_DIR)
	rm -rf $(BPF_OUT)
	rm -f coverage.out coverage.html

# ============================================================================
# Install tools
# ============================================================================

.PHONY: install-tools
install-tools: ## Install development tools
	hack/install/install-tools.sh

# ============================================================================
# Help
# ============================================================================

.PHONY: help
help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-24s\033[0m %s\n", $$1, $$2}'
