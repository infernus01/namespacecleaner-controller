# Basic Makefile for NamespaceCleaner CRD

.PHONY: install-kind apply-crd apply-cr setup clean

# Install kind cluster
install-kind:
	@echo "Creating kind cluster..."
	kind create cluster --name namespacecleaner-demo

# Apply the CRD
apply-crd:
	@echo "Applying CRD..."
	kubectl apply -f config/crd/namespacecleaners.yaml

# Apply the Custom Resource
apply-cr:
	@echo "Applying Custom Resource..."
	kubectl apply -f example-namespacecleaner.yaml

# Full setup: install kind + apply CRD + apply CR
setup: install-kind apply-crd apply-cr
	@echo "Setup complete!"
	@kubectl get namespacecleaners

# Run the controller
run-controller:
	@echo "Running the controller..."
	go run cmd/controller/main.go

# Create test namespaces for demonstration
create-test-namespaces:
	@echo "Creating test namespaces..."
	kubectl create namespace test-env-1 --dry-run=client -o yaml | kubectl label --local -f - environment=test -o yaml | kubectl apply -f -
	kubectl create namespace test-env-2 --dry-run=client -o yaml | kubectl label --local -f - environment=test -o yaml | kubectl apply -f -
	kubectl create namespace prod-env --dry-run=client -o yaml | kubectl label --local -f - environment=prod -o yaml | kubectl apply -f -

# Clean up
clean:
	@echo "Deleting kind cluster..."
	kind delete cluster --name namespacecleaner-demo 