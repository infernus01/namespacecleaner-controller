# NamespaceCleaner Controller - Phase 2

A basic loop-based Kubernetes controller that demonstrates how to watch and process custom resources.

## What it does

This simple controller:
- **Step 4**: Lists NamespaceCleaner custom resources using generated clients
- **Step 5**: Runs a reconcile loop every 30 seconds to process each NamespaceCleaner

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

## Example Output

```
🚀 NamespaceCleaner Controller Starting...
=== Step 4: Testing Client ===
📋 Listing NamespaceCleaner resources...
Found 1 NamespaceCleaner resources:
  🧹 Name: my-cleaner
     Schedule: 0 2 * * *
     Selector: {map[environment:test] [] [] []}

=== Step 5: Starting Loop-Based Reconciler ===
🔄 Starting reconcile loop (every 30 seconds)...

--- Reconcile Cycle Started ---
🧹 Processing NamespaceCleaner: my-cleaner
   Schedule: 0 2 * * *
   Looking for namespaces with labels: map[environment:test]
   📦 Found matching namespace: test-env-1
   ⚠️  Would delete namespace 'test-env-1' (not actually deleting for safety)
   📦 Found matching namespace: test-env-2
   ⚠️  Would delete namespace 'test-env-2' (not actually deleting for safety)
--- Reconcile Cycle Completed ---
```

## Key Features

- **Simple loop-based design**: No complex event watching, just periodic scanning
- **Safe by default**: Logs what it would do instead of actually deleting
- **Uses generated clients**: Demonstrates how to use the auto-generated clientset
- **Basic reconciliation**: Shows the core pattern of reading desired state and taking action

This is the foundation for understanding how Kubernetes controllers work before moving to more advanced patterns. 