/*
Copyright 2024.

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

package controller

import (
	"context"
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	hellov1alpha1 "github.com/supertren/operator/api/v1alpha1"
)

const (
	typeAvailableHelloApp = "Available"
)

// HelloAppReconciler reconciles a HelloApp object
type HelloAppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=hello.example.com,resources=helloapps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hello.example.com,resources=helloapps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hello.example.com,resources=helloapps/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

func (r *HelloAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the HelloApp instance
	helloApp := &hellov1alpha1.HelloApp{}
	err := r.Get(ctx, req.NamespacedName, helloApp)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("HelloApp resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get HelloApp")
		return ctrl.Result{}, err
	}

	// Set initial status condition if not set
	if helloApp.Status.Conditions == nil || len(helloApp.Status.Conditions) == 0 {
		meta.SetStatusCondition(&helloApp.Status.Conditions, metav1.Condition{
			Type:               typeAvailableHelloApp,
			Status:             metav1.ConditionUnknown,
			ObservedGeneration: helloApp.Generation,
			Reason:             "Reconciling",
			Message:            "Starting reconciliation",
		})
		if err = r.Status().Update(ctx, helloApp); err != nil {
			logger.Error(err, "Failed to update HelloApp status")
			return ctrl.Result{}, err
		}
		if err := r.Get(ctx, req.NamespacedName, helloApp); err != nil {
			logger.Error(err, "Failed to re-fetch HelloApp")
			return ctrl.Result{}, err
		}
	}

	// Define the desired Deployment object
	dep, err := r.deploymentForHelloApp(helloApp)
	if err != nil {
		logger.Error(err, "Failed to define new Deployment resource for HelloApp")
		meta.SetStatusCondition(&helloApp.Status.Conditions, metav1.Condition{
			Type:               typeAvailableHelloApp,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: helloApp.Generation,
			Reason:             "Reconciling",
			Message:            fmt.Sprintf("Failed to create Deployment: %s", err),
		})
		if err := r.Status().Update(ctx, helloApp); err != nil {
			logger.Error(err, "Failed to update HelloApp status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	}

	// Check if the Deployment already exists, if not create one
	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: helloApp.Name, Namespace: helloApp.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		logger.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		if err = r.Create(ctx, dep); err != nil {
			logger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	// Ensure the deployment size is the same as the spec
	needsUpdate := false
	// dep.Spec.Replicas is set by us when creating the desired Deployment.
	// The existing Deployment (`found`) may have a nil Replicas pointer, so
	// guard dereferences to avoid panics.
	if found.Spec.Replicas == nil && dep.Spec.Replicas != nil {
		found.Spec.Replicas = dep.Spec.Replicas
		needsUpdate = true
	} else if found.Spec.Replicas != nil && dep.Spec.Replicas != nil {
		if *found.Spec.Replicas != *dep.Spec.Replicas {
			found.Spec.Replicas = dep.Spec.Replicas
			needsUpdate = true
		}
	}

	// Ensure the container environment variables (Message) match
	if len(found.Spec.Template.Spec.Containers) > 0 && len(dep.Spec.Template.Spec.Containers) > 0 {
		if !reflect.DeepEqual(found.Spec.Template.Spec.Containers[0].Env, dep.Spec.Template.Spec.Containers[0].Env) {
			found.Spec.Template.Spec.Containers[0].Env = dep.Spec.Template.Spec.Containers[0].Env
			needsUpdate = true
		}
	}

	if needsUpdate {
		logger.Info("Updating Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
		if err = r.Update(ctx, found); err != nil {
			logger.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Update status with available replicas
	meta.SetStatusCondition(&helloApp.Status.Conditions, metav1.Condition{
		Type:               typeAvailableHelloApp,
		Status:             metav1.ConditionTrue,
		ObservedGeneration: helloApp.Generation,
		Reason:             "Reconciling",
		Message:            fmt.Sprintf("Deployment for HelloApp (%s) created successfully", helloApp.Name),
	})

	helloApp.Status.AvailableReplicas = found.Status.AvailableReplicas
	if err = r.Status().Update(ctx, helloApp); err != nil {
		logger.Error(err, "Failed to update HelloApp status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func boolPtr(b bool) *bool {
	return &b
}

func (r *HelloAppReconciler) deploymentForHelloApp(helloApp *hellov1alpha1.HelloApp) (*appsv1.Deployment, error) {
	replicas := helloApp.Spec.Replicas

	labels := map[string]string{
		"app":        helloApp.Name,
		"managed-by": "hello-operator",
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      helloApp.Name,
			Namespace: helloApp.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name: "hello-app",
						// nginxinc/nginx-unprivileged runs as non-root on port 8080 — required for OpenShift SCC
						Image: "nginxinc/nginx-unprivileged:alpine",
						Env: []corev1.EnvVar{
							{
								Name:  "HELLO_MESSAGE",
								Value: helloApp.Spec.Message,
							},
						},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Protocol:      corev1.ProtocolTCP,
						}},
						SecurityContext: &corev1.SecurityContext{
							AllowPrivilegeEscalation: boolPtr(false),
							RunAsNonRoot:             boolPtr(true),
							Capabilities: &corev1.Capabilities{
								Drop: []corev1.Capability{"ALL"},
							},
							SeccompProfile: &corev1.SeccompProfile{
								Type: corev1.SeccompProfileTypeRuntimeDefault,
							},
						},
					}},
				},
			},
		},
	}

	// Set HelloApp as the owner of the Deployment
	if err := ctrl.SetControllerReference(helloApp, dep, r.Scheme); err != nil {
		return nil, err
	}

	return dep, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HelloAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hellov1alpha1.HelloApp{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
