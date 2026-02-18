# Variables - set IMG to your registry image
IMG ?= controller:latest
NAMESPACE ?= hello-operator-system

# Tool versions
ENVTEST_K8S_VERSION = 1.29.0
CONTROLLER_TOOLS_VERSION ?= v0.14.0
KUSTOMIZE_VERSION ?= v5.3.0

# Get the currently used golang install path
GOBIN=$(shell go env GOPATH)/bin

.PHONY: all
all: build

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", $$1 } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code (deepcopy functions).
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet.
	go vet ./...

##@ Build

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

.PHONY: run
run: manifests generate fmt vet ## Run against the configured Kubernetes cluster.
	go run ./main.go

.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	docker build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG}

##@ Deployment

.PHONY: install
install: manifests ## Install CRDs into the cluster.
	kubectl apply -f config/crd/bases/

.PHONY: uninstall
uninstall: manifests ## Uninstall CRDs from the cluster.
	kubectl delete --ignore-not-found=true -f config/crd/bases/

.PHONY: deploy
deploy: manifests ## Deploy controller to the cluster.
	kubectl apply -f config/manager/manager.yaml
	kubectl apply -f config/rbac/service_account.yaml
	kubectl apply -f config/rbac/role.yaml
	kubectl apply -f config/rbac/role_binding.yaml
	kubectl set image deployment/hello-operator-controller-manager manager=${IMG} -n ${NAMESPACE}

.PHONY: undeploy
undeploy: ## Undeploy controller from the cluster.
	kubectl delete --ignore-not-found=true -f config/rbac/role_binding.yaml
	kubectl delete --ignore-not-found=true -f config/rbac/role.yaml
	kubectl delete --ignore-not-found=true -f config/rbac/service_account.yaml
	kubectl delete --ignore-not-found=true -f config/manager/manager.yaml

.PHONY: sample
sample: ## Apply the sample HelloApp CR.
	kubectl apply -f config/samples/hello_v1alpha1_helloapp.yaml

##@ Tools

CONTROLLER_GEN = $(GOBIN)/controller-gen
.PHONY: controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-install-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen,$(CONTROLLER_TOOLS_VERSION))

define go-install-tool
@[ -f $(1) ] || { \
set -e ;\
echo "Downloading $(2)@$(3)" ;\
GOBIN=$(GOBIN) go install $(2)@$(3) ;\
}
endef
