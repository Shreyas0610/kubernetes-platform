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
		})
	})
})
