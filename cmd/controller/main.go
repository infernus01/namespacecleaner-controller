/*
Simple pod cleanup controller:
- Reads NamespaceCleaner custom resources  
- Deletes old pods (older than 1 hour) in matching namespaces
- Runs every 30 seconds
*/

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
	fmt.Println("NamespaceCleaner Controller Starting...")

	// Create clients
	client, k8sClient, err := createClients()
	if err != nil {
		log.Fatalf("Failed to create clients: %v", err)
	}

	// Start the main loop
	startLoop(client, k8sClient)
}

// createClients creates both custom and standard Kubernetes clients
func createClients() (clientset.Interface, kubernetes.Interface, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first, then fall back to kubeconfig
	config, err = rest.InClusterConfig()
	if err != nil {
		kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create config: %v", err)
		}
	}

	// Create both clients
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

// startLoop runs the controller loop
func startLoop(client clientset.Interface, k8sClient kubernetes.Interface) {
	for {
		fmt.Println("--- Cleanup Cycle ---")

		// Get all NamespaceCleaner resources
		namespacecleaners, err := client.ClusteropsV1alpha1().NamespaceCleaners().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			log.Printf("Failed to list NamespaceCleaners: %v", err)
		} else {
			// Process each one
			for _, nc := range namespacecleaners.Items {
				cleanupPods(nc, k8sClient)
			}
		}

		time.Sleep(30 * time.Second)
	}
}

// cleanupPods finds and deletes old pods
func cleanupPods(nc v1alpha1.NamespaceCleaner, k8sClient kubernetes.Interface) {
	fmt.Printf("Processing: %s\n", nc.Name)

	if len(nc.Spec.Selector.MatchLabels) == 0 {
		fmt.Printf("No selector, skipping\n")
		return
	}

	// Get all namespaces
	namespaces, err := k8sClient.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("Failed to list namespaces: %v\n", err)
		return
	}

	totalDeleted := 0
	
	// Check each namespace
	for _, ns := range namespaces.Items {
		if matchesSelector(ns.Labels, nc.Spec.Selector.MatchLabels) {
			fmt.Printf("Found namespace: %s\n", ns.Name)
			
			// Get pods in this namespace
			pods, err := k8sClient.CoreV1().Pods(ns.Name).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				fmt.Printf("Failed to list pods: %v\n", err)
				continue
			}
			
			// Delete old pods (older than 1 hour)
			for _, pod := range pods.Items {
				age := time.Since(pod.CreationTimestamp.Time)
				if age > time.Hour {
					fmt.Printf("  Deleting old pod: %s\n", pod.Name)
					err := k8sClient.CoreV1().Pods(ns.Name).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
					if err != nil {
						fmt.Printf("  Failed: %v\n", err)
					} else {
						totalDeleted++
					}
				}
			}
		}
	}
	
	fmt.Printf("Deleted %d old pods\n", totalDeleted)
}

// matchesSelector checks if labels match the selector
func matchesSelector(labels map[string]string, selectorLabels map[string]string) bool {
	if labels == nil {
		return false
	}

	for key, value := range selectorLabels {
		if labels[key] != value {
			return false
		}
	}

	return true
}
