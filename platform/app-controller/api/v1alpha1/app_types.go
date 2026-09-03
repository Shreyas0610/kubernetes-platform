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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// AppHealthCheck defines HTTP health probe settings for an app container.
type AppHealthCheck struct {
	// Path is the HTTP path used for readiness and liveness checks.
	// +kubebuilder:validation:Pattern=`^/.*`
	// +optional
	Path string `json:"path,omitempty"`
}

// AppSpec defines the desired state of App.
type AppSpec struct {
	// Image is the container image deployed by the platform.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Image string `json:"image"`

	// Port is the container and service port exposed by the app.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`

	// Replicas is the desired number of app replicas.
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=1
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// Host is the optional HTTP hostname exposed through Ingress.
	// +optional
	Host string `json:"host,omitempty"`

	// TLS enables HTTPS routing for Host through cert-manager-managed Ingress TLS.
	// +optional
	TLS bool `json:"tls,omitempty"`

	// Env is non-sensitive runtime configuration stored in a generated ConfigMap.
	// +optional
	Env map[string]string `json:"env,omitempty"`

	// EnvFromConfigMap is the name of an existing ConfigMap loaded into the app container.
	// +optional
	EnvFromConfigMap string `json:"envFromConfigMap,omitempty"`

	// EnvFromSecret is the name of an existing Secret loaded into the app container.
	// +optional
	EnvFromSecret string `json:"envFromSecret,omitempty"`

	// HealthCheck configures HTTP readiness and liveness probes for the app container.
	// +optional
	HealthCheck *AppHealthCheck `json:"healthCheck,omitempty"`

	// Resources configures CPU and memory requests and limits for the app container.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// AppStatus defines the observed state of App.
type AppStatus struct {
	// Phase is a compact summary of the app lifecycle.
	// +optional
	Phase string `json:"phase,omitempty"`

	// URL is the externally reachable URL when host routing is configured.
	// +optional
	URL string `json:"url,omitempty"`

	// Conditions describe the observed reconciliation state.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// App is the Schema for the apps API
type App struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of App
	// +required
	Spec AppSpec `json:"spec"`

	// status defines the observed state of App
	// +optional
	Status AppStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// AppList contains a list of App
type AppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []App `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &App{}, &AppList{})
		return nil
	})
}
