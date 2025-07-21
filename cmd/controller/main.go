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

	// Create clients
	client, k8sClient, err := createClients()
	if err != nil {
		log.Fatalf("Failed to create clients: %v", err)
	}

	// listing NamespaceCleaner resources
	fmt.Println("=== Testing Client ===")
	err = listNamespaceCleaners(client)
	if err != nil {
		log.Fatalf("Failed to list NamespaceCleaners: %v", err)
	}

	// Starting the reconciler
	fmt.Println("=== Starting the Reconciler ===")
	startReconcileLoop(client, k8sClient)
}

// createClients creates both custom and standard Kubernetes clients
func createClients() (clientset.Interface, kubernetes.Interface, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first, then fall back to kubeconfig
	config, err = rest.InClusterConfig()
	if err != nil {
		// We're not in cluster, use kubeconfig
		kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create config: %v", err)
		}
	}

	// Create both clients using the same config
	customClient, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create custom clientset: %v", err)
	}

	k8sClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create kubernetes clientset: %v", err)
	}

	return customClient, k8sClient, nil
}

// listNamespaceCleaners to list our custom resources
func listNamespaceCleaners(client clientset.Interface) error {
	fmt.Println("Listing NamespaceCleaner resources...")

	// List all NamespaceCleaner resources (cluster-scoped)
	namespacecleaners, err := client.ClusteropsV1alpha1().NamespaceCleaners().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list NamespaceCleaners: %v", err)
	}

	fmt.Printf("Found %d NamespaceCleaner resources:\n", len(namespacecleaners.Items))

	for _, nc := range namespacecleaners.Items {
		fmt.Printf("     Name: %s\n", nc.Name)
		fmt.Printf("     Schedule: %s\n", nc.Spec.Schedule)
		fmt.Printf("     Selector: %+v\n", nc.Spec.Selector)
		fmt.Println()
	}

	return nil
}

// Simple loop-based reconciler
func startReconcileLoop(client clientset.Interface, k8sClient kubernetes.Interface) {
	fmt.Println("Starting reconcile loop (every 30 seconds)...")

	for {
		fmt.Println("\n--- Reconcillation Started ---")

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

		fmt.Println("--- Reconcillation Completed ---")

		// Waits 30 seconds before next cycle
		time.Sleep(30 * time.Second)
	}
}

// to process a single NamespaceCleaner resource
func reconcileNamespaceCleaner(nc v1alpha1.NamespaceCleaner, k8sClient kubernetes.Interface) error {

	fmt.Printf("Processing NamespaceCleaner: %s\n", nc.Name)
	fmt.Printf("Schedule: %s\n", nc.Spec.Schedule)

	// If selector has labels, find matching namespaces
	if len(nc.Spec.Selector.MatchLabels) > 0 {
		fmt.Printf("Looking for namespaces with labels: %v\n", nc.Spec.Selector.MatchLabels)

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
