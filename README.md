# NamespaceCleaner Controller

A basic loop-based Kubernetes controller that demonstrates how to watch and process custom resources.

## What it does

This simple controller:
- Lists NamespaceCleaner custom resources using generated clients
- Runs a reconcile loop every 30 seconds to process each NamespaceCleaner

## How to use

1. **Setup your cluster and CRD**:
   ```bash
   make setup
   ```

2. **Create test namespaces** (optional):
   ```bash
   make create-test-namespaces
   ```

3. **Run the controller**:
   ```bash
   make run-controller
   ```

## What you'll see

The controller will:
- List all NamespaceCleaner resources on startup
- Every 30 seconds, scan for NamespaceCleaner resources
- For each resource, find matching namespaces based on label selectors
- Log what it would delete (but won't actually delete for safety)
