# Basic Makefile for NamespaceCleaner CRD

.PHONY: install-kind apply-crd apply-cr setup clean

# Install kind cluster
install-kind:
	@echo "Creating kind cluster..."
	@kind create cluster --name namespacecleaner-demo

# Apply ALL CRDs (any YAML file in config/crd/)
apply-crds:
	@echo "Applying all CRDs..."
	@if [ -d "config/crd" ]; then \
		kubectl apply -f config/crd/; \
		echo "Applied all CRDs from config/crd/"; \
	else \
		echo "config/crd/ directory not found"; \
	fi

# Apply ALL Custom Resources (any YAML file in examples/)
apply-crs:
	@echo "Applying all Custom Resources..."
	@if [ -d "examples" ]; then \
		kubectl apply -f examples/; \
		echo "Applied all CRs from examples/"; \
	fi

# Full setup: install kind + apply ALL CRDs + apply ALL CRs
setup: install-kind apply-crds apply-crs
	@echo "Setup complete!"

# Run the controller
run-controller:
	@echo "Running the controller..."
	@go run cmd/controller/main.go

# Create test namespaces for demonstration
create-test-namespaces:
	@echo "Creating test namespaces..."
	@kubectl create namespace test-env-1 --dry-run=client -o yaml | kubectl label --local -f - environment=test -o yaml | kubectl apply -f -
	@kubectl create namespace test-env-2 --dry-run=client -o yaml | kubectl label --local -f - environment=staging -o yaml | kubectl apply -f -
	@kubectl create namespace prod-env --dry-run=client -o yaml | kubectl label --local -f - environment=prod -o yaml | kubectl apply -f -

# Clean up everything (cluster)
clean:
	@echo "Cleaning up..."
	@kubectl delete namespacecleaners --all || echo "No NamespaceCleaner resources to delete"
	@kind get clusters | grep -q namespacecleaner-demo && kind delete cluster --name namespacecleaner-demo || echo "Cluster 'namespacecleaner-demo' not found"