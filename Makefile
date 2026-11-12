.DEFAULT_GOAL := help

-include .env
export KUBE_CONTEXT

TOOLS_DIR := .tools

TOOL_OS := $(shell go env GOOS)
TOOL_ARCH := $(shell go env GOARCH)

GOVULNCHECK_VERSION := v1.9.0
GOLANGCI_LINT_VERSION := v2.15.0

GOVULNCHECK := $(TOOLS_DIR)/govulncheck-$(GOVULNCHECK_VERSION)-$(TOOL_OS)-$(TOOL_ARCH)
GOLANGCI_LINT := $(TOOLS_DIR)/golangci-lint-$(GOLANGCI_LINT_VERSION)-$(TOOL_OS)-$(TOOL_ARCH)
TOOL_INSTALL_DIR := $(TOOLS_DIR)/.install

.PHONY: help

##@ Help

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n"} /^##@/ {printf "\n\033[1m%s\033[0m\n", substr($$0, 5)} /^[a-zA-Z0-9_%-]+:.*##/ {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build test race vet fmt fmt-check fix

##@ Go

build: ## Build all packages
	go build -mod=readonly ./...

test: ## Run unit tests
	go test -mod=readonly ./...

race: ## Run tests with the race detector
	go test -mod=readonly -race ./...

vet: ## Run go vet
	go vet -mod=readonly ./...

fmt: ## Format Go source
	go fmt ./...

fmt-check: ## Verify Go source formatting
	@test -z "$$(gofmt -l $$(git ls-files --cached --others --exclude-standard -- '*.go'))"

fix: ## Apply standard Go fixes
	go fix ./...

.PHONY: mod-tidy mod-tidy-check mod-verify vuln lint ci

##@ Quality

mod-tidy: ## Synchronize go.mod and go.sum with the source tree
	go mod tidy

mod-tidy-check: ## Verify go.mod and go.sum need no changes
	go mod tidy -diff

mod-verify: ## Verify downloaded modules against go.sum
	go mod verify

vuln: tools ## Run govulncheck
	$(GOVULNCHECK) ./...

lint: tools ## Run golangci-lint
	$(GOLANGCI_LINT) run

ci: fmt-check mod-tidy-check mod-verify build vet race vuln lint ## Run the complete CI quality gate

.PHONY: dev-run dev-container-build dev-container-run dev-container-smoke edge-build edge-publish edge-deploy edge-rollout edge-verify edge-up

DEV_IMAGE ?= anya:dev
DEV_PORT ?= 8080
DEV_SMOKE_PORT ?= 0
DEV_PROVIDER ?= replay
##@ Development

dev-run: ## Run Anya locally with the development provider
	AGENT_PROVIDER="$(DEV_PROVIDER)" go run ./cmd/node-agent

dev-container-build: ## Build the local development container
	IMAGE="$(DEV_IMAGE)" \
		./hack/build-container.sh

dev-container-run: dev-container-build ## Run the local development container
	docker run --rm --publish "$(DEV_PORT):8080" \
		--env AGENT_PROVIDER="$(DEV_PROVIDER)" \
		"$(DEV_IMAGE)"

dev-container-smoke: dev-container-build ## Run the local container smoke test
	./hack/smoke-container.sh "$(DEV_IMAGE)" "$(DEV_SMOKE_PORT)"

.PHONY: edge-build edge-publish edge-deploy edge-rollout edge-verify edge-up

GIT_COMMIT := $(shell git rev-parse HEAD 2>/dev/null)
EDGE_IMAGE_REPOSITORY ?= ghcr.io/sojournerdev/anya
EDGE_IMAGE_TAG ?= git-$(GIT_COMMIT)
EDGE_IMAGE ?= $(EDGE_IMAGE_REPOSITORY):$(EDGE_IMAGE_TAG)
KUBE_IMAGE ?= $(EDGE_IMAGE)
EDGE_TARGET_PLATFORM ?= linux/amd64
KUBE_NAMESPACE ?= anya-system
KUBE_ROLLOUT_TIMEOUT ?= 120s
EDGE_SERVICE ?= anya-agent
EDGE_LOCAL_PORT ?= 8080
EDGE_REMOTE_PORT ?= 8080
EDGE_READY_TIMEOUT_SECONDS ?= 30

##@ Edge Deployment

edge-build: ## Build the versioned edge image
	IMAGE="$(EDGE_IMAGE)" \
	TARGET_PLATFORM="$(EDGE_TARGET_PLATFORM)" \
	IMAGE_REVISION="$(GIT_COMMIT)" \
		./hack/build-container.sh

edge-publish: edge-build ## Push the versioned edge image to its OCI registry
	docker push "$(EDGE_IMAGE)"

edge-deploy: ## Apply the edge Kubernetes resources
	KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	KUBE_IMAGE="$(KUBE_IMAGE)" \
	KUBE_NAMESPACE="$(KUBE_NAMESPACE)" \
		./hack/deploy-k8s.sh

edge-rollout: ## Wait for the edge rollout
	kubectl --context "$(KUBE_CONTEXT)" --namespace "$(KUBE_NAMESPACE)" \
		rollout status daemonset/anya --timeout="$(KUBE_ROLLOUT_TIMEOUT)"

edge-verify: ## Verify the edge deployment
	KUBE_CONTEXT="$(KUBE_CONTEXT)" \
	KUBE_NAMESPACE="$(KUBE_NAMESPACE)" \
	EDGE_SERVICE="$(EDGE_SERVICE)" \
	EDGE_LOCAL_PORT="$(EDGE_LOCAL_PORT)" \
	EDGE_REMOTE_PORT="$(EDGE_REMOTE_PORT)" \
	EDGE_READY_TIMEOUT_SECONDS="$(EDGE_READY_TIMEOUT_SECONDS)" \
		./hack/verify-service.sh

edge-up: ## Build, publish, deploy, and verify the edge workflow
	@set -eu; \
	$(MAKE) --no-print-directory edge-publish; \
	$(MAKE) --no-print-directory edge-deploy; \
	$(MAKE) --no-print-directory edge-rollout; \
	$(MAKE) --no-print-directory edge-verify

.PHONY: tools clean-tools

##@ Tools

tools: $(GOVULNCHECK) $(GOLANGCI_LINT) ## Install pinned local tools

$(GOVULNCHECK):
	mkdir -p "$(TOOL_INSTALL_DIR)"
	GOBIN="$(abspath $(TOOL_INSTALL_DIR))" go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	mv "$(TOOL_INSTALL_DIR)/govulncheck" "$@"
	rm -rf "$(TOOL_INSTALL_DIR)"

$(GOLANGCI_LINT):
	mkdir -p "$(TOOL_INSTALL_DIR)"
	GOBIN="$(abspath $(TOOL_INSTALL_DIR))" go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	mv "$(TOOL_INSTALL_DIR)/golangci-lint" "$@"
	rm -rf "$(TOOL_INSTALL_DIR)"

clean-tools: ## Remove local tools
	rm -rf $(TOOLS_DIR)
