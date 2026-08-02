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
	"k8s.io/apimachinery/pkg/api/errors"
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
					// TODO(user): Specify other spec details if needed.
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
			// TODO(user): Add more specific assertions depending on your controller's reconciliation logic.
			// Example: If you expect a certain status condition after reconciliation, verify it here.
		})
	})
})
