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
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	platformv1alpha1 "github.com/sarigeb/kubernetes-platform/platform/app-controller/api/v1alpha1"
)

func TestLabelsForAppIncludesPlatformLabels(t *testing.T) {
	app := &platformv1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{Name: "demo-api"},
	}

	labels := labelsForApp(app)

	if labels["app.kubernetes.io/name"] != "demo-api" {
		t.Fatalf("expected app name label demo-api, got %q", labels["app.kubernetes.io/name"])
	}
	if labels["app.kubernetes.io/managed-by"] != "kubernetes-platform" {
		t.Fatalf("expected managed-by label kubernetes-platform, got %q", labels["app.kubernetes.io/managed-by"])
	}
	if labels["platform.sarige.dev/app"] != "demo-api" {
		t.Fatalf("expected platform app label demo-api, got %q", labels["platform.sarige.dev/app"])
	}
}

func TestReplicasForAppDefaultsToOne(t *testing.T) {
	app := &platformv1alpha1.App{}

	replicas := replicasForApp(app)

	if replicas != 1 {
		t.Fatalf("expected default replicas 1, got %d", replicas)
	}
}

func TestReplicasForAppUsesSpecValue(t *testing.T) {
	var desired int32 = 3
	app := &platformv1alpha1.App{
		Spec: platformv1alpha1.AppSpec{Replicas: &desired},
	}

	replicas := replicasForApp(app)

	if replicas != 3 {
		t.Fatalf("expected replicas 3, got %d", replicas)
	}
}

func TestDeploymentForAppBuildsExpectedDeployment(t *testing.T) {
	var desired int32 = 2
	app := &platformv1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-api",
			Namespace: "default",
		},
		Spec: platformv1alpha1.AppSpec{
			Image:    "nginx:1.27",
			Port:     80,
			Replicas: &desired,
		},
	}
	scheme := runtime.NewScheme()
	if err := platformv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add platform scheme: %v", err)
	}
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatalf("add apps scheme: %v", err)
	}

	deployment, err := deploymentForApp(app, scheme)
	if err != nil {
		t.Fatalf("deploymentForApp returned error: %v", err)
	}

	if deployment.Name != "demo-api" {
		t.Fatalf("expected deployment name demo-api, got %q", deployment.Name)
	}
	if *deployment.Spec.Replicas != 2 {
		t.Fatalf("expected 2 replicas, got %d", *deployment.Spec.Replicas)
	}
	container := deployment.Spec.Template.Spec.Containers[0]
	if container.Image != "nginx:1.27" {
		t.Fatalf("expected image nginx:1.27, got %q", container.Image)
	}
	if container.Ports[0].ContainerPort != 80 {
		t.Fatalf("expected container port 80, got %d", container.Ports[0].ContainerPort)
	}
	if len(deployment.OwnerReferences) != 1 || deployment.OwnerReferences[0].Name != "demo-api" {
		t.Fatalf("expected owner reference to demo-api, got %#v", deployment.OwnerReferences)
	}
}

func TestDeploymentForAppInjectsRuntimeConfiguration(t *testing.T) {
	app := &platformv1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-api",
			Namespace: "default",
		},
		Spec: platformv1alpha1.AppSpec{
			Image:            "nginx:1.27",
			Port:             80,
			Env:              map[string]string{"LOG_LEVEL": "debug"},
			EnvFromConfigMap: "shared-app-config",
			EnvFromSecret:    "demo-api-secrets",
		},
	}
	scheme := runtime.NewScheme()
	if err := platformv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add platform scheme: %v", err)
	}
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatalf("add apps scheme: %v", err)
	}

	deployment, err := deploymentForApp(app, scheme)
	if err != nil {
		t.Fatalf("deploymentForApp returned error: %v", err)
	}

	envFrom := deployment.Spec.Template.Spec.Containers[0].EnvFrom
	if len(envFrom) != 3 {
		t.Fatalf("expected three envFrom sources, got %#v", envFrom)
	}
	if envFrom[0].ConfigMapRef.Name != "demo-api-env" {
		t.Fatalf("expected generated ConfigMap source demo-api-env, got %#v", envFrom[0])
	}
	if envFrom[1].ConfigMapRef.Name != "shared-app-config" {
		t.Fatalf("expected external ConfigMap source shared-app-config, got %#v", envFrom[1])
	}
	if envFrom[2].SecretRef.Name != "demo-api-secrets" {
		t.Fatalf("expected Secret source demo-api-secrets, got %#v", envFrom[2])
	}
}

func TestDeploymentForAppConfiguresHTTPHealthProbes(t *testing.T) {
	app := &platformv1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-api",
			Namespace: "default",
		},
		Spec: platformv1alpha1.AppSpec{
			Image: "nginx:1.27",
			Port:  8080,
			HealthCheck: &platformv1alpha1.AppHealthCheck{
				Path: "/health",
			},
		},
	}
	scheme := runtime.NewScheme()
	if err := platformv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add platform scheme: %v", err)
	}
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatalf("add apps scheme: %v", err)
	}

	deployment, err := deploymentForApp(app, scheme)
	if err != nil {
		t.Fatalf("deploymentForApp returned error: %v", err)
	}

	container := deployment.Spec.Template.Spec.Containers[0]
	if container.ReadinessProbe == nil {
		t.Fatal("expected readiness probe")
	}
	if container.ReadinessProbe.HTTPGet.Path != "/health" {
		t.Fatalf("expected readiness path /health, got %q", container.ReadinessProbe.HTTPGet.Path)
	}
	if container.ReadinessProbe.HTTPGet.Port.StrVal != "http" {
		t.Fatalf("expected readiness probe to target named http port, got %#v", container.ReadinessProbe.HTTPGet.Port)
	}
	if container.LivenessProbe == nil {
		t.Fatal("expected liveness probe")
	}
	if container.LivenessProbe.HTTPGet.Path != "/health" {
		t.Fatalf("expected liveness path /health, got %q", container.LivenessProbe.HTTPGet.Path)
	}
}

func TestDeploymentForAppConfiguresResourceRequirements(t *testing.T) {
	app := &platformv1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-api",
			Namespace: "default",
		},
		Spec: platformv1alpha1.AppSpec{
			Image: "nginx:1.27",
			Port:  80,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("500m"),
					corev1.ResourceMemory: resource.MustParse("512Mi"),
				},
			},
		},
	}
	scheme := runtime.NewScheme()
	if err := platformv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add platform scheme: %v", err)
	}
	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatalf("add apps scheme: %v", err)
	}

	deployment, err := deploymentForApp(app, scheme)
	if err != nil {
		t.Fatalf("deploymentForApp returned error: %v", err)
	}

	resources := deployment.Spec.Template.Spec.Containers[0].Resources
	if resources.Requests.Cpu().String() != "100m" {
		t.Fatalf("expected cpu request 100m, got %s", resources.Requests.Cpu().String())
	}
	if resources.Requests.Memory().String() != "128Mi" {
		t.Fatalf("expected memory request 128Mi, got %s", resources.Requests.Memory().String())
	}
	if resources.Limits.Cpu().String() != "500m" {
		t.Fatalf("expected cpu limit 500m, got %s", resources.Limits.Cpu().String())
	}
	if resources.Limits.Memory().String() != "512Mi" {
		t.Fatalf("expected memory limit 512Mi, got %s", resources.Limits.Memory().String())
	}
}

func TestConfigMapForAppBuildsExpectedConfigMap(t *testing.T) {
	app := &platformv1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-api",
			Namespace: "default",
		},
		Spec: platformv1alpha1.AppSpec{
			Env: map[string]string{
				"LOG_LEVEL": "debug",
				"REGION":    "local",
			},
		},
	}
	scheme := runtime.NewScheme()
	if err := platformv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add platform scheme: %v", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("add core scheme: %v", err)
	}

	configMap, err := configMapForApp(app, scheme)
	if err != nil {
		t.Fatalf("configMapForApp returned error: %v", err)
	}

	if configMap.Name != "demo-api-env" {
		t.Fatalf("expected ConfigMap name demo-api-env, got %q", configMap.Name)
	}
	if configMap.Data["LOG_LEVEL"] != "debug" || configMap.Data["REGION"] != "local" {
		t.Fatalf("unexpected ConfigMap data: %#v", configMap.Data)
	}
	if len(configMap.OwnerReferences) != 1 || configMap.OwnerReferences[0].Name != "demo-api" {
		t.Fatalf("expected owner reference to demo-api, got %#v", configMap.OwnerReferences)
	}
}

func TestServiceForAppBuildsExpectedService(t *testing.T) {
	app := &platformv1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-api",
			Namespace: "default",
		},
		Spec: platformv1alpha1.AppSpec{
			Image: "nginx:1.27",
			Port:  80,
		},
	}
	scheme := runtime.NewScheme()
	if err := platformv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add platform scheme: %v", err)
	}
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("add core scheme: %v", err)
	}

	service, err := serviceForApp(app, scheme)
	if err != nil {
		t.Fatalf("serviceForApp returned error: %v", err)
	}

	if service.Name != "demo-api" {
		t.Fatalf("expected service name demo-api, got %q", service.Name)
	}
	if service.Spec.Type != corev1.ServiceTypeClusterIP {
		t.Fatalf("expected ClusterIP service, got %q", service.Spec.Type)
	}
	if service.Spec.Ports[0].Port != 80 {
		t.Fatalf("expected service port 80, got %d", service.Spec.Ports[0].Port)
	}
	if service.Spec.Selector["platform.sarige.dev/app"] != "demo-api" {
		t.Fatalf("expected selector for demo-api, got %#v", service.Spec.Selector)
	}
}

func TestIngressForAppBuildsExpectedIngress(t *testing.T) {
	app := &platformv1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-api",
			Namespace: "default",
		},
		Spec: platformv1alpha1.AppSpec{
			Image: "nginx:1.27",
			Port:  80,
			Host:  "demo.local",
		},
	}
	scheme := runtime.NewScheme()
	if err := platformv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add platform scheme: %v", err)
	}
	if err := networkingv1.AddToScheme(scheme); err != nil {
		t.Fatalf("add networking scheme: %v", err)
	}

	ingress, err := ingressForApp(app, scheme)
	if err != nil {
		t.Fatalf("ingressForApp returned error: %v", err)
	}

	if ingress.Name != "demo-api" {
		t.Fatalf("expected ingress name demo-api, got %q", ingress.Name)
	}
	if ingress.Spec.Rules[0].Host != "demo.local" {
		t.Fatalf("expected host demo.local, got %q", ingress.Spec.Rules[0].Host)
	}
	if ingress.Spec.IngressClassName == nil || *ingress.Spec.IngressClassName != "nginx" {
		t.Fatalf("expected ingress class nginx, got %v", ingress.Spec.IngressClassName)
	}
	path := ingress.Spec.Rules[0].HTTP.Paths[0]
	if path.Backend.Service.Name != "demo-api" {
		t.Fatalf("expected backend service demo-api, got %q", path.Backend.Service.Name)
	}
	if path.Backend.Service.Port.Number != 80 {
		t.Fatalf("expected backend service port 80, got %d", path.Backend.Service.Port.Number)
	}
}

func TestIngressForAppAddsCertManagerTLSWhenEnabled(t *testing.T) {
	app := &platformv1alpha1.App{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-api",
			Namespace: "default",
		},
		Spec: platformv1alpha1.AppSpec{
			Image: "nginx:1.27",
			Port:  80,
			Host:  "demo.local",
			TLS:   true,
		},
	}
	scheme := runtime.NewScheme()
	if err := platformv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add platform scheme: %v", err)
	}
	if err := networkingv1.AddToScheme(scheme); err != nil {
		t.Fatalf("add networking scheme: %v", err)
	}

	ingress, err := ingressForApp(app, scheme)
	if err != nil {
		t.Fatalf("ingressForApp returned error: %v", err)
	}

	if ingress.Annotations["cert-manager.io/cluster-issuer"] != "platform-local-selfsigned" {
		t.Fatalf("expected cert-manager cluster issuer annotation, got %#v", ingress.Annotations)
	}
	if len(ingress.Spec.TLS) != 1 {
		t.Fatalf("expected one TLS entry, got %d", len(ingress.Spec.TLS))
	}
	if ingress.Spec.TLS[0].SecretName != "demo-api-tls" {
		t.Fatalf("expected TLS secret demo-api-tls, got %q", ingress.Spec.TLS[0].SecretName)
	}
	if len(ingress.Spec.TLS[0].Hosts) != 1 || ingress.Spec.TLS[0].Hosts[0] != "demo.local" {
		t.Fatalf("expected TLS host demo.local, got %#v", ingress.Spec.TLS[0].Hosts)
	}
}

func TestURLForAppUsesHTTPSWhenTLSEnabled(t *testing.T) {
	app := &platformv1alpha1.App{
		Spec: platformv1alpha1.AppSpec{
			Host: "demo.local",
			TLS:  true,
		},
	}

	url := urlForApp(app)

	if url != "https://demo.local" {
		t.Fatalf("expected https URL, got %q", url)
	}
}

func TestSetAppConditionReplacesExistingCondition(t *testing.T) {
	app := &platformv1alpha1.App{
		Status: platformv1alpha1.AppStatus{
			Conditions: []metav1.Condition{{
				Type:   "Ready",
				Status: metav1.ConditionFalse,
				Reason: "OldReason",
			}},
		},
	}

	setAppCondition(app, metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
		Reason: "Reconciled",
	})

	if len(app.Status.Conditions) != 1 {
		t.Fatalf("expected one Ready condition, got %d", len(app.Status.Conditions))
	}
	if app.Status.Conditions[0].Status != metav1.ConditionTrue {
		t.Fatalf("expected Ready=True, got %s", app.Status.Conditions[0].Status)
	}
	if app.Status.Conditions[0].Reason != "Reconciled" {
		t.Fatalf("expected Reconciled reason, got %q", app.Status.Conditions[0].Reason)
	}
}

func TestRuntimeStatusReadyWhenDeploymentHasDesiredAvailableReplicas(t *testing.T) {
	var desired int32 = 2
	app := &platformv1alpha1.App{
		Spec: platformv1alpha1.AppSpec{Replicas: &desired},
	}
	deployment := &appsv1.Deployment{
		Status: appsv1.DeploymentStatus{AvailableReplicas: 2},
	}

	status := runtimeStatusForApp(app, deployment)

	if status.phase != "Ready" {
		t.Fatalf("expected Ready phase, got %q", status.phase)
	}
	assertCondition(t, status.conditions, "Ready", metav1.ConditionTrue, "Available")
	assertCondition(t, status.conditions, "Reconciling", metav1.ConditionFalse, "Available")
	assertCondition(t, status.conditions, "Stalled", metav1.ConditionFalse, "Available")
}

func TestRuntimeStatusReconcilingWhenDeploymentIsStillRollingOut(t *testing.T) {
	var desired int32 = 3
	app := &platformv1alpha1.App{
		Spec: platformv1alpha1.AppSpec{Replicas: &desired},
	}
	deployment := &appsv1.Deployment{
		Status: appsv1.DeploymentStatus{AvailableReplicas: 1},
	}

	status := runtimeStatusForApp(app, deployment)

	if status.phase != "Reconciling" {
		t.Fatalf("expected Reconciling phase, got %q", status.phase)
	}
	assertCondition(t, status.conditions, "Ready", metav1.ConditionFalse, "WaitingForDeployment")
	assertCondition(t, status.conditions, "Reconciling", metav1.ConditionTrue, "WaitingForDeployment")
	assertCondition(t, status.conditions, "Stalled", metav1.ConditionFalse, "WaitingForDeployment")
}

func TestRuntimeStatusStalledWhenDeploymentProgressDeadlineExceeded(t *testing.T) {
	app := &platformv1alpha1.App{}
	deployment := &appsv1.Deployment{
		Status: appsv1.DeploymentStatus{
			Conditions: []appsv1.DeploymentCondition{{
				Type:   appsv1.DeploymentProgressing,
				Status: corev1.ConditionFalse,
				Reason: "ProgressDeadlineExceeded",
			}},
		},
	}

	status := runtimeStatusForApp(app, deployment)

	if status.phase != "Stalled" {
		t.Fatalf("expected Stalled phase, got %q", status.phase)
	}
	assertCondition(t, status.conditions, "Ready", metav1.ConditionFalse, "ProgressDeadlineExceeded")
	assertCondition(t, status.conditions, "Reconciling", metav1.ConditionFalse, "ProgressDeadlineExceeded")
	assertCondition(t, status.conditions, "Stalled", metav1.ConditionTrue, "ProgressDeadlineExceeded")
}

func assertCondition(t *testing.T, conditions []metav1.Condition, conditionType string, status metav1.ConditionStatus, reason string) {
	t.Helper()
	condition := findCondition(conditions, conditionType)
	if condition == nil {
		t.Fatalf("expected condition %q, got %#v", conditionType, conditions)
	}
	if condition.Status != status {
		t.Fatalf("expected %s=%s, got %s", conditionType, status, condition.Status)
	}
	if condition.Reason != reason {
		t.Fatalf("expected %s reason %q, got %q", conditionType, reason, condition.Reason)
	}
}

func findCondition(conditions []metav1.Condition, conditionType string) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}

var _ = Describe("App Controller", func() {
	Context("When reconciling a resource", func() {
		const (
			resourceName      = "test-resource"
			resourceNamespace = "default"
		)

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: resourceNamespace,
		}
		app := &platformv1alpha1.App{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind App")
			err := k8sClient.Get(ctx, typeNamespacedName, app)
			if err != nil && errors.IsNotFound(err) {
				resource := &platformv1alpha1.App{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: resourceNamespace,
					},
					Spec: platformv1alpha1.AppSpec{
						Image: "nginx:1.27",
						Port:  80,
						Host:  "test.local",
						TLS:   true,
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &platformv1alpha1.App{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance App")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &AppReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("updating App status with an HTTPS URL")
			reconciled := &platformv1alpha1.App{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, reconciled)).To(Succeed())
			Expect(reconciled.Status.URL).To(Equal("https://test.local"))
		})
	})
})
