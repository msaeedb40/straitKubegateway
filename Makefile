# straitKubegateway Makefile
# ──────────────────────────────────────────────────────────────────────────────

SHELL           := /bin/bash
GO              := go
GOARCH          ?= $(shell $(GO) env GOARCH)
GOOS            ?= linux
GO_VERSION      := 1.26.5
MODULE          := github.com/straitKubegateway/straitKubegateway

# Image configuration
REGISTRY        ?= ghcr.io/straitkubegateway
IMAGE_TAG       ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
PLATFORMS       ?= linux/amd64,linux/arm64

# BPF toolchain
CLANG           ?= clang
LLC             ?= llc
LLVM_STRIP      ?= llvm-strip
BPF_CFLAGS      := -O2 -g -Wall -Werror -target bpf
BPF_INCLUDES    := -I bpf/include

# Directories
BIN_DIR         := bin
BPF_DIR         := bpf
BUILD_DIR       := build
HACK_DIR        := hack
PROTO_DIR       := proto

# Binaries
STRAITD         := $(BIN_DIR)/straitd
SG_CONTROLLER   := $(BIN_DIR)/sg-controller
SG_CLI          := $(BIN_DIR)/sg-cli
CNI_PLUGIN      := $(BIN_DIR)/straitkubegateway-cni

# Go build flags
LDFLAGS         := -s -w \
                   -X $(MODULE)/internal/version.Version=$(IMAGE_TAG) \
                   -X $(MODULE)/internal/version.GitCommit=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
                   -X $(MODULE)/internal/version.BuildDate=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# ──────────────────────────────────────────────────────────────────────────────
# Default target
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: all
all: build

# ──────────────────────────────────────────────────────────────────────────────
# Build
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: build build-straitd build-sg-controller build-sg-cli build-cni build-bpf

build: build-bpf build-straitd build-sg-controller build-sg-cli build-cni

build-straitd:
	@echo "Building straitd ($(GOOS)/$(GOARCH))..."
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build \
		-ldflags "$(LDFLAGS)" \
		-o $(STRAITD) \
		./cmd/straitd

build-sg-controller:
	@echo "Building sg-controller ($(GOOS)/$(GOARCH))..."
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build \
		-ldflags "$(LDFLAGS)" \
		-o $(SG_CONTROLLER) \
		./cmd/sg-controller

build-sg-cli:
	@echo "Building sg-cli ($(GOOS)/$(GOARCH))..."
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build \
		-ldflags "$(LDFLAGS)" \
		-o $(SG_CLI) \
		./cmd/sg-cli

build-cni:
	@echo "Building CNI plugin ($(GOOS)/$(GOARCH))..."
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build \
		-ldflags "$(LDFLAGS)" \
		-o $(CNI_PLUGIN) \
		./cni/plugin

build-bpf:
	@echo "Building BPF programs..."
	@mkdir -p $(BIN_DIR)/bpf
	@for src in $(wildcard $(BPF_DIR)/*/*.c); do \
		obj=$$(echo $$src | sed 's|$(BPF_DIR)/|$(BIN_DIR)/bpf/|;s|\.c$$|.o|'); \
		dir=$$(dirname $$obj); \
		mkdir -p $$dir; \
		echo "  CC $$src -> $$obj"; \
		$(CLANG) $(BPF_CFLAGS) $(BPF_INCLUDES) -c $$src -o $$obj; \
		$(LLVM_STRIP) -g $$obj 2>/dev/null || true; \
	done

# ──────────────────────────────────────────────────────────────────────────────
# Container images
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: build-images build-image-straitd build-image-controller build-image-cli build-image-ui

build-images: build-image-straitd build-image-controller build-image-cli build-image-ui

build-image-straitd:
	@echo "Building straitd image..."
	docker buildx build \
		--platform $(PLATFORMS) \
		-t $(REGISTRY)/straitd:$(IMAGE_TAG) \
		-f $(BUILD_DIR)/Dockerfile.straitd \
		--push .

build-image-controller:
	@echo "Building sg-controller image..."
	docker buildx build \
		--platform $(PLATFORMS) \
		-t $(REGISTRY)/sg-controller:$(IMAGE_TAG) \
		-f $(BUILD_DIR)/Dockerfile.sg-controller \
		--push .

build-image-cli:
	@echo "Building sg-cli image..."
	docker buildx build \
		--platform $(PLATFORMS) \
		-t $(REGISTRY)/sg-cli:$(IMAGE_TAG) \
		-f $(BUILD_DIR)/Dockerfile.sg-cli \
		--push .

build-image-ui:
	@echo "Building UI image..."
	docker buildx build \
		--platform $(PLATFORMS) \
		-t $(REGISTRY)/ui:$(IMAGE_TAG) \
		-f $(BUILD_DIR)/Dockerfile.ui \
		--push .

# ──────────────────────────────────────────────────────────────────────────────
# Code generation
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: generate generate-crds generate-deepcopy generate-client generate-bpf proto

generate: generate-deepcopy generate-crds generate-client generate-bpf proto

generate-crds:
	@echo "Generating CRDs..."
	$(GO) run sigs.k8s.io/controller-tools/cmd/controller-gen \
		crd paths="./api/..." output:crd:artifacts:config=straitKubegateway-helm/charts/crd/templates

generate-deepcopy:
	@echo "Generating deepcopy..."
	$(GO) run sigs.k8s.io/controller-tools/cmd/controller-gen \
		object:headerFile="$(HACK_DIR)/boilerplate/license_header.txt" paths="./api/..."

generate-client:
	@echo "Generating client..."
	@bash $(HACK_DIR)/generate/generate-client.sh

generate-bpf:
	@echo "Generating BPF Go bindings..."
	$(GO) generate ./ebpf/...

proto:
	@echo "Generating protobuf..."
	buf generate $(PROTO_DIR)

# ──────────────────────────────────────────────────────────────────────────────
# Testing
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: test test-unit test-integration test-e2e test-dataplane test-ui

test: test-unit

test-unit:
	@echo "Running unit tests..."
	$(GO) test -race -count=1 -coverprofile=coverage.out ./...

test-integration:
	@echo "Running integration tests..."
	$(GO) test -race -tags=integration -count=1 ./test/integration/...

test-e2e:
	@echo "Running e2e tests..."
	$(GO) test -timeout 30m -tags=e2e -count=1 ./test/e2e/...

test-dataplane:
	@echo "Running dataplane tests..."
	$(GO) test -race -tags=dataplane -count=1 ./test/dataplane/...

test-ui:
	@echo "Running UI tests..."
	cd ui && npm test

# ──────────────────────────────────────────────────────────────────────────────
# Linting & verification
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: lint verify verify-gofmt verify-govet verify-generated verify-crds verify-bpf verify-manifests

lint:
	@echo "Running linters..."
	golangci-lint run ./...

verify: verify-gofmt verify-govet verify-generated verify-crds verify-bpf verify-manifests

verify-gofmt:
	@bash $(HACK_DIR)/verify/verify-gofmt.sh

verify-govet:
	@bash $(HACK_DIR)/verify/verify-govet.sh

verify-generated:
	@bash $(HACK_DIR)/verify/verify-generated.sh

verify-crds:
	@bash $(HACK_DIR)/verify/verify-crds.sh

verify-bpf:
	@bash $(HACK_DIR)/verify/verify-bpf.sh

verify-manifests:
	@bash $(HACK_DIR)/verify/verify-manifests.sh

# ──────────────────────────────────────────────────────────────────────────────
# Install
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: install install-tools install-clang install-cni

install: install-tools

install-tools:
	@bash $(HACK_DIR)/install/install-tools.sh

install-clang:
	@bash $(HACK_DIR)/install/install-clang.sh

install-cni:
	@bash $(HACK_DIR)/install/install-cni.sh

# ──────────────────────────────────────────────────────────────────────────────
# Helm
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: helm-lint helm-template helm-package

helm-lint:
	helm lint straitKubegateway-helm/

helm-template:
	helm template straitkubegateway straitKubegateway-helm/ --debug

helm-package:
	helm package straitKubegateway-helm/ -d $(BUILD_DIR)/helm

# ──────────────────────────────────────────────────────────────────────────────
# Clean
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: clean

clean:
	@echo "Cleaning..."
	rm -rf $(BIN_DIR) coverage.out
	cd ui && rm -rf dist node_modules/.cache

# ──────────────────────────────────────────────────────────────────────────────
# Help
# ──────────────────────────────────────────────────────────────────────────────

.PHONY: help

help:
	@echo "straitKubegateway Makefile"
	@echo ""
	@echo "Build:"
	@echo "  make build              Build all binaries"
	@echo "  make build-straitd      Build node agent"
	@echo "  make build-sg-controller Build control plane"
	@echo "  make build-sg-cli       Build CLI tool"
	@echo "  make build-cni          Build CNI plugin"
	@echo "  make build-bpf          Build BPF programs"
	@echo "  make build-images       Build all container images"
	@echo ""
	@echo "Generate:"
	@echo "  make generate           Run all code generation"
	@echo "  make generate-crds      Generate CRD manifests"
	@echo "  make generate-deepcopy  Generate deepcopy methods"
	@echo "  make proto              Generate protobuf code"
	@echo ""
	@echo "Test:"
	@echo "  make test               Run unit tests"
	@echo "  make test-integration   Run integration tests"
	@echo "  make test-e2e           Run end-to-end tests"
	@echo "  make test-dataplane     Run dataplane tests"
	@echo "  make test-ui            Run Angular UI tests"
	@echo ""
	@echo "Other:"
	@echo "  make lint               Run linters"
	@echo "  make verify             Run all verifications"
	@echo "  make helm-lint          Lint Helm chart"
	@echo "  make clean              Clean build artifacts"
