/*
Copyright 2024 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package namespacecleaner

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/pkg/reconciler"

	"github.com/infernus01/knative-demo/pkg/apis/clusterops/v1alpha1"
	namespacecleanerlister "github.com/infernus01/knative-demo/pkg/generated/listers/clusterops/v1alpha1"
)

// Reconciler implements controller.Reconciler for NamespaceCleaner resources.
type Reconciler struct {
	kubeclientset          kubernetes.Interface
	namespacecleanerLister namespacecleanerlister.NamespaceCleanerLister
}

// Check that our Reconciler implements Interface
var _ controller.Reconciler = (*Reconciler)(nil)

// Check that our Reconciler implements LeaderAware
var _ reconciler.LeaderAware = (*Reconciler)(nil)

// Reconcile implements controller.Reconciler
func (r *Reconciler) Reconcile(ctx context.Context, key string) error {
	logger := logging.FromContext(ctx).With(zap.String("namespacecleaner", key))
	logger.Info("Reconciling NamespaceCleaner")

	// Get the NamespaceCleaner resource with this name
	namespaceCleaner, err := r.namespacecleanerLister.Get(key)
	if errors.IsNotFound(err) {
		// The NamespaceCleaner resource may no longer exist, in which case we stop processing.
		logger.Info("NamespaceCleaner resource no longer exists")
		return nil
	} else if err != nil {
		return err
	}

	return r.reconcileNamespaceCleaner(ctx, namespaceCleaner)
}

func (r *Reconciler) reconcileNamespaceCleaner(ctx context.Context, nc *v1alpha1.NamespaceCleaner) error {
	logger := logging.FromContext(ctx).With(zap.String("namespacecleaner", nc.Name))

	// Check if selector is specified
	if len(nc.Spec.Selector.MatchLabels) == 0 {
		logger.Info("No selector specified, skipping cleanup")
		return nil
	}

	// List all namespaces
	namespaces, err := r.kubeclientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list namespaces: %w", err)
	}

	totalDeleted := 0

	for _, ns := range namespaces.Items {
		// Skip system namespaces (kube-* prefixed)
		if strings.HasPrefix(ns.Name, "kube-") {
			logger.Debugw("Skipping system namespace", zap.String("namespace", ns.Name))
			continue
		}

		// Check if namespace matches selector
		if r.matchesSelector(ns.Labels, nc.Spec.Selector.MatchLabels) {
			logger.Infow("Processing namespace", zap.String("namespace", ns.Name))

			deleted, err := r.cleanupOldPods(ctx, ns.Name)
			if err != nil {
				logger.Errorw("Error cleaning namespace",
					zap.String("namespace", ns.Name),
					zap.Error(err))
				// Continue with other namespaces even if one fails
				continue
			}
			totalDeleted += deleted
		}
	}

	logger.Infow("Cleanup completed",
		zap.String("namespacecleaner", nc.Name),
		zap.Int("totalDeleted", totalDeleted))

	return nil
}

func (r *Reconciler) cleanupOldPods(ctx context.Context, namespace string) (int, error) {
	logger := logging.FromContext(ctx).With(zap.String("namespace", namespace))

	pods, err := r.kubeclientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to list pods in namespace %s: %w", namespace, err)
	}

	deleted := 0
	cutoff := time.Now().Add(-30 * time.Second) // Demo: Delete pods older than 30 seconds

	for _, pod := range pods.Items {
		// Only delete completed pods (Succeeded or Failed)
		if pod.Status.Phase == "Succeeded" || pod.Status.Phase == "Failed" {
			if pod.CreationTimestamp.Time.Before(cutoff) {
				logger.Infow("Deleting old pod (older than 30 seconds)",
					zap.String("pod", pod.Name),
					zap.String("phase", string(pod.Status.Phase)),
					zap.Duration("age", time.Since(pod.CreationTimestamp.Time)))

				err := r.kubeclientset.CoreV1().Pods(namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
				if err != nil {
					logger.Errorw("Failed to delete pod",
						zap.String("pod", pod.Name),
						zap.Error(err))
					continue
				}
				deleted++
			}
		}
	}

	return deleted, nil
}

func (r *Reconciler) matchesSelector(labels map[string]string, selectorLabels map[string]string) bool {
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

// Promote implements reconciler.LeaderAware
func (r *Reconciler) Promote(bkt reconciler.Bucket, enq func(reconciler.Bucket, types.NamespacedName)) error {
	// This is called when we become the leader.
	// For this simple controller, we don't need to do anything special.
	return nil
}

// Demote implements reconciler.LeaderAware
func (r *Reconciler) Demote(bkt reconciler.Bucket) {
	// This is called when we are no longer the leader.
	// For this simple controller, we don't need to do anything special.
}
