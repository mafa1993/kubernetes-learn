/*
Copyright 2017 The Kubernetes Authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// App is a specification for a App resource
type App struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec AppSpec `json:"spec"`
}

// region -----需要自己修改的----
// AppSpec is the spec for a App resource
type AppSpec struct {
	Deployment DeploymentSpec `json:"deployment"`
	Service    ServiceSpec    `json:"service"`
	Ingress    IngressSpec    `json:"ingress"`
}

// DeploymentSpec 定义deployment  里面的属性
type DeploymentSpec struct {
	Image    string `json:"image"`    // 使用的镜像
	Replicas *int32 `json:"replicas"` // 副本数
	Name     string `json:"name"`     // 名字
}

type ServiceSpec struct {
	Name       string `json:"name"`       // svc 名
	Port       uint16 `json:"port"`       // svc 端口，默认是cluster Ip模式
	TargetPort uint16 `json:"targetport"` // 后端服务的端口
}

type IngressSpec struct {
	Name string `json:"name"` // ingress 名
	Path string `json:"path"` // ingress 访问的路径
}

// endregion

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppList is a list of App resources
type AppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []App `json:"items"`
}
