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
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	platformv1alpha1 "github.com/sarigeb/kubernetes-platform/platform/app-controller/api/v1alpha1"
)

const (
	certManagerClusterIssuerAnnotation = "cert-manager.io/cluster-issuer"
	localClusterIssuerName             = "platform-local-selfsigned"
	nginxIngressClassName              = "nginx"
)

type appRuntimeStatus struct {
	phase      string
	conditions []metav1.Condition
}

// AppReconciler reconciles a App object
type AppReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=platform.sarige.dev,resources=apps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=platform.sarige.dev,resources=apps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=platform.sarige.dev,resources=apps/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps;services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

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

	if len(app.Spec.Env) > 0 {
		desiredConfigMap, err := configMapForApp(&app, r.Scheme)
		if err != nil {
			return ctrl.Result{}, err
		}
		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      desiredConfigMap.Name,
				Namespace: desiredConfigMap.Namespace,
			},
		}
		_, err = controllerutil.CreateOrPatch(ctx, r.Client, configMap, func() error {
			configMap.Labels = desiredConfigMap.Labels
			configMap.OwnerReferences = desiredConfigMap.OwnerReferences
			configMap.Data = desiredConfigMap.Data
			return nil
		})
		if err != nil {
			return ctrl.Result{}, err
		}
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

	if app.Spec.Host != "" {
		desiredIngress, err := ingressForApp(&app, r.Scheme)
		if err != nil {
			return ctrl.Result{}, err
		}
		ingress := &networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      desiredIngress.Name,
				Namespace: desiredIngress.Namespace,
			},
		}
		_, err = controllerutil.CreateOrPatch(ctx, r.Client, ingress, func() error {
			ingress.Annotations = desiredIngress.Annotations
			ingress.Labels = desiredIngress.Labels
			ingress.OwnerReferences = desiredIngress.OwnerReferences
			ingress.Spec = desiredIngress.Spec
			return nil
		})
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.updateReadyStatus(ctx, req.NamespacedName); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *AppReconciler) updateReadyStatus(ctx context.Context, name types.NamespacedName) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var app platformv1alpha1.App
		if err := r.Get(ctx, name, &app); err != nil {
			return err
		}
		var deployment appsv1.Deployment
		if err := r.Get(ctx, name, &deployment); err != nil {
			return err
		}

		runtimeStatus := runtimeStatusForApp(&app, &deployment)
		app.Status.Phase = runtimeStatus.phase
		app.Status.URL = urlForApp(&app)
		for _, condition := range runtimeStatus.conditions {
			condition.ObservedGeneration = app.Generation
			setAppCondition(&app, condition)
		}
		return r.Status().Update(ctx, &app)
	})
}

func runtimeStatusForApp(app *platformv1alpha1.App, deployment *appsv1.Deployment) appRuntimeStatus {
	if deploymentProgressDeadlineExceeded(deployment) {
		return appRuntimeStatus{
			phase: "Stalled",
			conditions: []metav1.Condition{
				{Type: "Ready", Status: metav1.ConditionFalse, Reason: "ProgressDeadlineExceeded", Message: "Deployment exceeded its progress deadline"},
				{Type: "Reconciling", Status: metav1.ConditionFalse, Reason: "ProgressDeadlineExceeded", Message: "Deployment rollout is stalled"},
				{Type: "Stalled", Status: metav1.ConditionTrue, Reason: "ProgressDeadlineExceeded", Message: "Deployment exceeded its progress deadline"},
			},
		}
	}

	desiredReplicas := replicasForApp(app)
	if deployment.Status.AvailableReplicas >= desiredReplicas {
		return appRuntimeStatus{
			phase: "Ready",
			conditions: []metav1.Condition{
				{Type: "Ready", Status: metav1.ConditionTrue, Reason: "Available", Message: "Deployment has the desired number of available replicas"},
				{Type: "Reconciling", Status: metav1.ConditionFalse, Reason: "Available", Message: "Deployment rollout is complete"},
				{Type: "Stalled", Status: metav1.ConditionFalse, Reason: "Available", Message: "Deployment is not stalled"},
			},
		}
	}

	return appRuntimeStatus{
		phase: "Reconciling",
		conditions: []metav1.Condition{
			{Type: "Ready", Status: metav1.ConditionFalse, Reason: "WaitingForDeployment", Message: "Deployment does not have the desired number of available replicas"},
			{Type: "Reconciling", Status: metav1.ConditionTrue, Reason: "WaitingForDeployment", Message: "Deployment rollout is still progressing"},
			{Type: "Stalled", Status: metav1.ConditionFalse, Reason: "WaitingForDeployment", Message: "Deployment is not stalled"},
		},
	}
}

func deploymentProgressDeadlineExceeded(deployment *appsv1.Deployment) bool {
	for _, condition := range deployment.Status.Conditions {
		if condition.Type == appsv1.DeploymentProgressing &&
			condition.Status == corev1.ConditionFalse &&
			condition.Reason == "ProgressDeadlineExceeded" {
			return true
		}
	}
	return false
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
	envFrom := envFromSourcesForApp(app)
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
						Name:    app.Name,
						Image:   app.Spec.Image,
						EnvFrom: envFrom,
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

func envFromSourcesForApp(app *platformv1alpha1.App) []corev1.EnvFromSource {
	envFrom := []corev1.EnvFromSource{}
	if len(app.Spec.Env) > 0 {
		envFrom = append(envFrom, corev1.EnvFromSource{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: app.Name + "-env"},
			},
		})
	}
	if app.Spec.EnvFromConfigMap != "" {
		envFrom = append(envFrom, corev1.EnvFromSource{
			ConfigMapRef: &corev1.ConfigMapEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: app.Spec.EnvFromConfigMap},
			},
		})
	}
	if app.Spec.EnvFromSecret != "" {
		envFrom = append(envFrom, corev1.EnvFromSource{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{Name: app.Spec.EnvFromSecret},
			},
		})
	}
	return envFrom
}

func configMapForApp(app *platformv1alpha1.App, scheme *runtime.Scheme) (*corev1.ConfigMap, error) {
	labels := labelsForApp(app)
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name + "-env",
			Namespace: app.Namespace,
			Labels:    labels,
		},
		Data: app.Spec.Env,
	}
	if err := controllerutil.SetControllerReference(app, configMap, scheme); err != nil {
		return nil, err
	}
	return configMap, nil
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

func ingressForApp(app *platformv1alpha1.App, scheme *runtime.Scheme) (*networkingv1.Ingress, error) {
	pathType := networkingv1.PathTypePrefix
	ingressClassName := nginxIngressClassName
	labels := labelsForApp(app)
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			Labels:    labels,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &ingressClassName,
			Rules: []networkingv1.IngressRule{{
				Host: app.Spec.Host,
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{{
							Path:     "/",
							PathType: &pathType,
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: app.Name,
									Port: networkingv1.ServiceBackendPort{Number: app.Spec.Port},
								},
							},
						}},
					},
				},
			}},
		},
	}
	if app.Spec.TLS {
		ingress.Annotations = map[string]string{
			certManagerClusterIssuerAnnotation: localClusterIssuerName,
		}
		ingress.Spec.TLS = []networkingv1.IngressTLS{{
			Hosts:      []string{app.Spec.Host},
			SecretName: app.Name + "-tls",
		}}
	}
	if err := controllerutil.SetControllerReference(app, ingress, scheme); err != nil {
		return nil, err
	}
	return ingress, nil
}

func urlForApp(app *platformv1alpha1.App) string {
	if app.Spec.Host == "" {
		return ""
	}
	if app.Spec.TLS {
		return "https://" + app.Spec.Host
	}
	return "http://" + app.Spec.Host
}

func setAppCondition(app *platformv1alpha1.App, condition metav1.Condition) {
	meta.SetStatusCondition(&app.Status.Conditions, condition)
}

// SetupWithManager sets up the controller with the Manager.
func (r *AppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&platformv1alpha1.App{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&networkingv1.Ingress{}).
		Named("app").
		Complete(r)
}
