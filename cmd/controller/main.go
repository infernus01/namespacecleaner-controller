package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/infernus01/knative-demo/pkg/apis/clusterops/v1alpha1"
	clientset "github.com/infernus01/knative-demo/pkg/generated/clientset/versioned"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	fmt.Println("NamespaceCleaner Controller Starting...\n")

	// Step 4: Create a client to list resources
	client, err := createClient()
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Also create standard Kubernetes client for namespace operations
	k8sClient, err := createK8sClient()
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Test listing NamespaceCleaner resources (Step 4)
	fmt.Println("=== Step 4: Testing Client ===")
	err = listNamespaceCleaners(client)
	if err != nil {
		log.Fatalf("Failed to list NamespaceCleaners: %v", err)
	}

	// Step 5: Start the loop-based reconciler
	fmt.Println("=== Step 5: Starting Loop-Based Reconciler ===")
	startReconcileLoop(client, k8sClient)
}

// createClient creates a Kubernetes client using kubeconfig
func createClient() (clientset.Interface, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first, then fall back to kubeconfig
	config, err = rest.InClusterConfig()
	if err != nil {
		// We're not in cluster, use kubeconfig
		kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create config: %v", err)
		}
	}

	// Create the clientset using our generated client
	client, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %v", err)
	}

	return client, nil
}

// createK8sClient creates a standard Kubernetes client for namespace operations
func createK8sClient() (kubernetes.Interface, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first, then fall back to kubeconfig
	config, err = rest.InClusterConfig()
	if err != nil {
		// We're not in cluster, use kubeconfig
		kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create config: %v", err)
		}
	}

	// Create the standard Kubernetes clientset
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %v", err)
	}

	return client, nil
}

// listNamespaceCleaners demonstrates using the client to list our custom resources
func listNamespaceCleaners(client clientset.Interface) error {
	fmt.Println("📋 Listing NamespaceCleaner resources...")

	// List all NamespaceCleaner resources (cluster-scoped)
	namespacecleaners, err := client.ClusteropsV1alpha1().NamespaceCleaners().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list NamespaceCleaners: %v", err)
	}

	fmt.Printf("Found %d NamespaceCleaner resources:\n", len(namespacecleaners.Items))

	// Print details of each resource
	for _, nc := range namespacecleaners.Items {
		fmt.Printf("  🧹 Name: %s\n", nc.Name)
		fmt.Printf("     Schedule: %s\n", nc.Spec.Schedule)
		fmt.Printf("     Selector: %+v\n", nc.Spec.Selector)
		fmt.Println()
	}

	return nil
}

// startReconcileLoop implements Step 5 - a simple loop-based reconciler
func startReconcileLoop(client clientset.Interface, k8sClient kubernetes.Interface) {
	fmt.Println("Starting reconcile loop (every 30 seconds)...")

	for {
		fmt.Println("\n--- Reconcile Cycle Started ---")

		// Get all NamespaceCleaner resources
		namespacecleaners, err := client.ClusteropsV1alpha1().NamespaceCleaners().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Printf("Failed to list NamespaceCleaners: %v", err)
		} else {
			// Process each NamespaceCleaner resource
			for _, nc := range namespacecleaners.Items {
				err := reconcileNamespaceCleaner(nc, k8sClient)
				if err != nil {
					log.Printf("Failed to reconcile %s: %v", nc.Name, err)
				}
			}
		}

		fmt.Println("--- Reconcile Cycle Completed ---")

		// Wait 30 seconds before next cycle
		time.Sleep(30 * time.Second)
	}
}

// reconcileNamespaceCleaner processes a single NamespaceCleaner resource
func reconcileNamespaceCleaner(nc v1alpha1.NamespaceCleaner, k8sClient kubernetes.Interface) error {

	fmt.Printf("Processing NamespaceCleaner: %s\n", nc.Name)
	fmt.Printf("Schedule: %s\n", nc.Spec.Schedule)

	// Simple logic: if selector has labels, find matching namespaces
	if len(nc.Spec.Selector.MatchLabels) > 0 {
		fmt.Printf("Looking for namespaces with labels: %+v\n", nc.Spec.Selector.MatchLabels)

		// List all namespaces
		namespaces, err := k8sClient.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list namespaces: %v", err)
		}

		// Check each namespace against the selector
		matchingCount := 0
		for _, ns := range namespaces.Items {
			if matchesSelector(ns.Labels, nc.Spec.Selector.MatchLabels) {
				matchingCount++
				fmt.Printf("Found matching namespace: %s\n", ns.Name)

				// Delete the namespace
				fmt.Printf("Deleting namespace '%s'...\n", ns.Name)
				err := k8sClient.CoreV1().Namespaces().Delete(context.TODO(), ns.Name, metav1.DeleteOptions{})
				if err != nil {
					fmt.Printf("Failed to delete namespace '%s': %v\n", ns.Name, err)
				} else {
					fmt.Printf("Successfully deleted namespace '%s'\n", ns.Name)
				}
			}
		}

		if matchingCount == 0 {
			fmt.Printf("No matching namespaces found\n")
		}
	} else {
		fmt.Printf("No selector labels specified, skipping\n")
	}

	return nil
}

// matchesSelector checks if namespace labels match the selector
func matchesSelector(nsLabels map[string]string, selectorLabels map[string]string) bool {
	if nsLabels == nil {
		return false
	}

	// All selector labels must match
	for key, value := range selectorLabels {
		if nsLabels[key] != value {
			return false
		}
	}

	return true
}
