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

package v1

import (
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type MyDeployment struct {
	// +kubebuilder:validation:Required
	Image string `json:"image"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Maximum=8
	Replace int `json:"replace"`
}

// Replace 这个有一个问题，就是hpa最大8，这里设置超过8 的时候，就会一直导致更新，但是受限扩容不了，一直在刷日志

type MyService struct {
	// Type     string `json:"type,omitempty"`
	// +kubebuilder:validation:Optional
	Port int `json:"port"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Maximum=37000
	// +kubebuilder:validation:Minimum=30000
	NodePort int `json:"nodePort,omitempty"`
}

type MyIngress struct {
	// +kubebuilder:validation:Optional
	IsEnable bool `json:"isEnable,omitempty"`
	// +kubebuilder:validation:Optional
	Host string `json:"host,omitempty"`
	// +kubebuilder:validation:Optional
	Path string `json:"path,omitempty"`
}

// AppSpec defines the desired state of App
type AppSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Foo is an example field of App. Edit app_types.go to remove/update
	// Foo string `json:"foo,omitempty"`
	Deployment MyDeployment `json:"deployment"`
	Service    MyService    `json:"service"`
	Ingress    MyIngress    `json:"ingress,omitempty"`
}

// AppStatus defines the observed state of App
type AppStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// HealthyReplicas *int32 `json:"healthy_replicas,omitempty"`
	DeploymentStatus              appsv1.DeploymentStatus                     `json:"deploymentStatus"`
	ServiceSpec                   corev1.ServiceSpec                          `json:"service_spec"`
	IngressSpec                   netv1.IngressSpec                           `json:"ingress_spec"`
	HorizontalPodAutoscalerStatus autoscalingv2.HorizontalPodAutoscalerStatus `json:"horizontal_pod_autoscaler_status"`
	Selector                      string                                      `json:"selector"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// https://cloud.tencent.com/developer/article/1749750

// App is the Schema for the apps API
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Image",type="string",JSONPath=".spec.deployment.image",description="The Docker Image of MyAPP"
// +kubebuilder:printcolumn:name="Size",type="integer",JSONPath=".status.deploymentStatus.readyReplicas",description="Replicas of deploy"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.deployment.replicas,statuspath=.status.deploymentStatus.replicas,selectorpath=.status.selector
// +kubebuilder:resource:path=apps,categories=all,singular=aloys
type App struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppSpec   `json:"spec,omitempty"`
	Status AppStatus `json:"status,omitempty"`
}

// AppList contains a list of App
type AppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []App `json:"items"`
}

func init() {
	SchemeBuilder.Register(&App{}, &AppList{})
}
