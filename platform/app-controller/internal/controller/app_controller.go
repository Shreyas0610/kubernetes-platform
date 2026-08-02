/*
Copyright 2026.

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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	platformv1alpha1 "github.com/sarigeb/kubernetes-platform/platform/app-controller/api/v1alpha1"
)

// AppReconciler reconciles a App object
type AppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=platform.sarige.dev,resources=apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=platform.sarige.dev,resources=apps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=platform.sarige.dev,resources=apps/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the App object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.24.1/pkg/reconcile
func (r *AppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	var app platformv1alpha1.App
	if err := r.Get(ctx, req.NamespacedName, &app); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	desiredDeployment, err := deploymentForApp(&app, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      desiredDeployment.Name,
			Namespace: desiredDeployment.Namespace,
		},
	}
	_, err = controllerutil.CreateOrPatch(ctx, r.Client, deployment, func() error {
		deployment.Labels = desiredDeployment.Labels
		deployment.OwnerReferences = desiredDeployment.OwnerReferences
		deployment.Spec = desiredDeployment.Spec
		return nil
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	desiredService, err := serviceForApp(&app, r.Scheme)
	if err != nil {
		return ctrl.Result{}, err
	}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      desiredService.Name,
			Namespace: desiredService.Namespace,
		},
	}
	_, err = controllerutil.CreateOrPatch(ctx, r.Client, service, func() error {
		service.Labels = desiredService.Labels
		service.OwnerReferences = desiredService.OwnerReferences
		service.Spec.Type = desiredService.Spec.Type
		service.Spec.Selector = desiredService.Spec.Selector
		service.Spec.Ports = desiredService.Spec.Ports
		return nil
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func labelsForApp(app *platformv1alpha1.App) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       app.Name,
		"app.kubernetes.io/managed-by": "kubernetes-platform",
		"platform.sarige.dev/app":      app.Name,
	}
}

func replicasForApp(app *platformv1alpha1.App) int32 {
	if app.Spec.Replicas == nil {
		return 1
	}
	return *app.Spec.Replicas
}

func deploymentForApp(app *platformv1alpha1.App, scheme *runtime.Scheme) (*appsv1.Deployment, error) {
	labels := labelsForApp(app)
	replicas := replicasForApp(app)
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: labels},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  app.Name,
						Image: app.Spec.Image,
						Ports: []corev1.ContainerPort{{
							Name:          "http",
							ContainerPort: app.Spec.Port,
						}},
					}},
				},
			},
		},
	}
	if err := controllerutil.SetControllerReference(app, deployment, scheme); err != nil {
		return nil, err
	}
	return deployment, nil
}

func serviceForApp(app *platformv1alpha1.App, scheme *runtime.Scheme) (*corev1.Service, error) {
	labels := labelsForApp(app)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Name:       "http",
				Port:       app.Spec.Port,
				TargetPort: intstr.FromString("http"),
			}},
		},
	}
	if err := controllerutil.SetControllerReference(app, service, scheme); err != nil {
		return nil, err
	}
	return service, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&platformv1alpha1.App{}).
		Named("app").
		Complete(r)
}
